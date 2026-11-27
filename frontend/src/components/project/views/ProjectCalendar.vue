<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useQueryClient, useMutation} from '@tanstack/vue-query'
import FullCalendar from '@fullcalendar/vue3'
import dayGridPlugin from '@fullcalendar/daygrid'
import interactionPlugin from '@fullcalendar/interaction'
import calendarLocales from '@fullcalendar/core/locales-all'
import type {CalendarOptions, EventInput, EventDropArg} from '@fullcalendar/core'
import type {EventResizeDoneArg} from '@fullcalendar/interaction'
import {allTasksQuery, type TaskResponse} from '@/client/queries/tasks'
import {updateTaskMutationOptions} from '@/client/queries/taskMutations'
import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import AddTask from '@/components/tasks/AddTask.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import Modal from '@/components/misc/Modal.vue'
import XButton from '@/components/input/Button.vue'
import type {ProjectResponse} from '@/client/queries/projects'
const props = defineProps<{projectId:number,viewId:number,isLoadingProject?:boolean}>()
const {t, locale} = useI18n({useScope: 'global'})
const router = useRouter(), queryClient = useQueryClient()
const calendar = ref<InstanceType<typeof FullCalendar>>()
const tasks = ref<TaskResponse[]>([]), loading = ref(false), error = ref('')
const createOpen = ref(false), createDay = ref(localDay(new Date()))
const createProject = ref<ProjectResponse|null>(null)
const createProjectId = computed(()=>props.projectId > 0 ? props.projectId : createProject.value?.id ?? 0)
const createDueDate = computed(()=>{
	const [year,month,day]=createDay.value.split('-').map(Number)
	return new Date(year,month-1,day,23,59,59).toISOString()
})
function openCreate(day=localDay(new Date())) {createDay.value=day;createOpen.value=true}
async function tasksAdded() {createOpen.value=false;await loadTasks()}
const mutation = useMutation(updateTaskMutationOptions())
let request = 0
function date(value?:string) {
	if(!value||value.startsWith('0001-')) return null
	const parsed=new Date(value)
	return Number.isFinite(parsed.getTime())?parsed:null
}
function sameInstant(a:Date,b:Date) {return Math.floor(a.getTime()/1000)===Math.floor(b.getTime()/1000)}
function schedule(task:TaskResponse) {
	const start=date(task.start_date), end=date(task.end_date), due=date(task.due_date)
	return {start:start??due,end:start&&end&&end>start?end:null,due}
}
function localDay(d:Date) { return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}` }
// Calendar-only compact label; keep the persisted title and detail popup unchanged.
function calendarTitle(title:string) { return title.replace(/\s*·\s*\d{1,2}\.\d{1,2}(?:\.\d{4})?\s*·\s*/g,' · ') }
const events=computed<EventInput[]>(()=>tasks.value.flatMap(task=>{
	const {start,end,due}=schedule(task)
	if(!start) return []
	// Display dates only; keep the original instants on extendedProps.task.
	const allDay=true
	const color=/^[\da-f]{6}$/i.test(task.hex_color)?`#${task.hex_color}`:'#2563eb'
	const exclusiveEnd=end?localDay(new Date(end.getFullYear(),end.getMonth(),end.getDate()+1)):undefined
	const base:EventInput={id:String(task.id),title:task.title,start:localDay(start),end:exclusiveEnd??undefined,allDay,backgroundColor:color,borderColor:color,classNames:task.done?['calendar-done']:[],extendedProps:{task}}
	const result=[base]
	// Planned time and deadline are independent: retain a separate deadline marker.
	if(due&&date(task.start_date)&&localDay(due)!==localDay(end??start)&&(!date(task.end_date)||!sameInstant(due,date(task.end_date)!))) result.push({...base,id:`${task.id}-due`,title:`◆ ${task.title}`,start:localDay(due),end:undefined,allDay:true,editable:false})
	return result
}))
async function loadTasks() {
	const id=++request; loading.value=true;error.value=''
	try {
		// Every page, including intervals spanning the visible range.
		const result=await queryClient.fetchQuery({...allTasksQuery({project:props.projectId,view:props.viewId}),staleTime:0})
		if(id===request) tasks.value=result
	} catch { if(id===request) error.value=t('project.calendar.loadError') }
	finally { if(id===request) loading.value=false }
}
async function moveEvent(info:EventDropArg|EventResizeDoneArg) {
	const task=info.event.extendedProps.task as TaskResponse, old=schedule(task), start=info.event.start
	if(!start||!old.start) {info.revert();return}
	// Shift calendar days, never replace persisted times with display midnight.
	const dayDelta=(a:Date,b:Date)=>(Date.UTC(a.getFullYear(),a.getMonth(),a.getDate())-Date.UTC(b.getFullYear(),b.getMonth(),b.getDate()))/86400000
	const shift=(value:Date,days:number)=>{const moved=new Date(value);moved.setDate(moved.getDate()+days);return moved.toISOString()}
	const delta=dayDelta(start,old.start)
	// Existing mutation sends JSON PATCH for dates only, preserving task fields and relations.
	const changes:{id:number,start_date?:string,end_date?:string,due_date?:string}={id:task.id}
	if(date(task.start_date)) changes.start_date=shift(old.start,delta)
	else changes.due_date=shift(old.start,delta)
	const persistedEnd=date(task.end_date)
	if('startDelta' in info && !persistedEnd && info.event.end) {
		// Resizing a deadline-only task creates an explicit planned interval.
		const inclusiveEnd=new Date(info.event.end)
		inclusiveEnd.setDate(inclusiveEnd.getDate()-1)
		changes.start_date=shift(old.start,delta)
		changes.end_date=shift(old.start,dayDelta(inclusiveEnd,old.start))
	}
	if(persistedEnd) {
		const exclusiveEnd=new Date(persistedEnd.getFullYear(),persistedEnd.getMonth(),persistedEnd.getDate()+1)
		changes.end_date=shift(persistedEnd,info.event.end?dayDelta(info.event.end,exclusiveEnd):delta)
	}
	if(old.due&&date(task.start_date)) changes.due_date=persistedEnd&&sameInstant(old.due,persistedEnd)&&changes.end_date?changes.end_date:shift(old.due,delta)
	try {await mutation.mutateAsync(changes);await loadTasks()}
	catch {info.revert();error.value=t('project.calendar.saveError')}
}
const options=computed<CalendarOptions>(()=>({
	plugins:[dayGridPlugin,interactionPlugin],locales:calendarLocales,locale:locale.value,initialView:'dayGridMonth',firstDay:1,
	headerToolbar:{left:'today prev,next',center:'title',right:'dayGridMonth,dayGridWeek,dayGridDay'},

	height:'auto',dayMaxEvents:3,views:{dayGridWeek:{dayMaxEvents:false},dayGridDay:{dayMaxEvents:false}},fixedWeekCount:false,editable:!mutation.isPending.value,eventResizableFromStart:true,
	allDayMaintainDuration:true,moreLinkClick:()=>{
		requestAnimationFrame(()=>{
			const root=calendar.value?.$el as HTMLElement|undefined
			const popup=root?.querySelector<HTMLElement>('.fc-popover'), viewport=root?.querySelector<HTMLElement>('.fc-view-harness')
			if(popup&&viewport){const box=popup.getBoundingClientRect(),bounds=viewport.getBoundingClientRect();if(box.bottom>bounds.bottom)popup.style.top=`${popup.offsetTop-(box.bottom-bounds.bottom)-8}px`;if(box.right>bounds.right)popup.style.left=`${popup.offsetLeft-(box.right-bounds.right)-8}px`;if(box.left<bounds.left)popup.style.left=`${popup.offsetLeft+(bounds.left-box.left)+8}px`}
		})
		return 'popover'
	},displayEventTime:false,eventDisplay:'block',
	events:events.value,eventDrop:moveEvent,eventResize:moveEvent,eventClick:info=>{openTask(info.event.extendedProps.task)},
	eventDidMount:info=>{info.el.title=info.event.title;info.el.setAttribute('aria-label',info.event.title);info.el.tabIndex=0;info.el.setAttribute('data-task-id',String(info.event.extendedProps.task.id));info.el.addEventListener('keydown',event=>{if(event.key==='Enter'||event.key===' '){event.preventDefault();openTask(info.event.extendedProps.task)}})},
	selectable:true,selectMinDistance:8,dateClick:info=>openCreate(localDay(info.date)),
	select:info=>{openCreate(localDay(info.start));calendar.value?.getApi().unselect()},
}))
function openTask(task:TaskResponse) {
	router.push({name:'task.detail',params:{id:task.id},state:{backdropView:router.currentRoute.value.fullPath}})
}
watch(()=>[props.projectId,props.viewId],loadTasks,{immediate:true})
</script>

<template>
	<ProjectWrapper
		:project-id="projectId"
		:view-id="viewId"
		:is-loading-project="isLoadingProject??false"
		class="calendar-project-wrapper"
	>
		<section
			class="project-calendar"
			:aria-label="t('project.calendar.title')"
		>
			<div class="calendar-caption">
				<strong>{{ t('project.calendar.title') }}</strong><span>{{ loading?t('project.calendar.loading'):t('project.calendar.count',{count:tasks.length}) }} · ◆ {{ t('project.calendar.deadline') }}</span><XButton
					icon="plus"
					@click="openCreate()"
				>
					{{ t('project.calendar.create') }}
				</XButton>
			</div>
			<Modal
				:enabled="createOpen"
				@close="createOpen=false"
			>
				<section
					class="calendar-create"
					:aria-label="t('project.calendar.create')"
				>
					<h2>{{ t('project.calendar.create') }}</h2>
					<label>{{ t('project.calendar.date') }} <input
						v-model="createDay"
						class="input"
						type="date"
						required
					></label>
					<ProjectSearch
						v-if="projectId<=0"
						v-model="createProject"
						:filter="p=>p.id>0&&!p.is_archived"
					/>
					<AddTask
						v-if="createDay&&createProjectId>0"
						:project-id="createProjectId"
						:default-due-date="createDueDate"
						@tasksAdded="tasksAdded"
					/>
					<p v-else>
						{{ t('project.calendar.chooseProject') }}
					</p>
					<XButton
						variant="secondary"
						@click="createOpen=false"
					>
						{{ t('project.calendar.cancel') }}
					</XButton>
				</section>
			</Modal>
			<div
				v-if="error"
				class="calendar-error"
				role="alert"
			>
				{{ error }} <button @click="loadTasks">
					{{ t('project.calendar.retry') }}
				</button>
			</div>
			<FullCalendar
				ref="calendar"
				:options="options"
			>
				<template #eventContent="arg">
					<div class="calendar-event-content">
						<span class="calendar-event-title">{{ calendarTitle(arg.event.title) }}</span>
					</div>
				</template>
			</FullCalendar>
			<p class="calendar-hint">
				{{ t('project.calendar.hint') }}
			</p>
		</section>
	</ProjectWrapper>
</template>

<style scoped>
.calendar-create{
color:var(--text);
background:var(--white);
border-radius:12px;
padding:24px;
min-inline-size:min(620px,80vw);
display:grid;
gap:18px
}

.calendar-project-wrapper{
inline-size:100%;
max-inline-size:none
}

.project-calendar{
background:var(--white);
border:1px solid var(--border);
border-radius:16px;
padding:20px;
min-inline-size:0;

--fc-border-color:var(--border);
--fc-page-bg-color:var(--white);
--fc-neutral-bg-color:var(--grey-100);
--fc-today-bg-color:#4285f40b;
--fc-button-bg-color:var(--white);
--fc-button-text-color:var(--text);
--fc-button-border-color:var(--border);
--fc-button-hover-bg-color:var(--grey-100);
--fc-button-hover-border-color:var(--grey-300);
--fc-button-active-bg-color:#2563eb;
--fc-button-active-border-color:#2563eb
}

.calendar-caption{
align-items:center;
flex-wrap:wrap;
display:flex;
justify-content:space-between;
gap:12px;
color:var(--text-light);
font-size:12px;
margin-block-end:18px
}

.project-calendar :deep(.fc-toolbar-title){
font-size:24px;
font-weight:500;
text-transform:capitalize
}

.project-calendar :deep(.fc-button){
font-size:13px;
padding:8px 12px;
box-shadow:none;
text-transform:capitalize
}

.project-calendar :deep(.fc-button-active){
color:white!important
}

.project-calendar :deep(.fc-scrollgrid){
border-radius:10px;
overflow:hidden
}

.project-calendar :deep(.fc-col-header-cell){
padding-block:10px;
font-size:12px;
font-weight:500;
color:var(--text-light)
}

.project-calendar :deep(.fc-dayGridMonth-view .fc-daygrid-day-frame){
min-block-size:115px
}

.project-calendar :deep(.fc-view-harness){
overflow-x:auto;
overscroll-behavior-inline:contain
}

.project-calendar :deep(.fc-dayGridWeek-view){
min-inline-size:840px
}

.project-calendar :deep(.fc-dayGridWeek-view .fc-daygrid-day-frame),.project-calendar :deep(.fc-dayGridDay-view .fc-daygrid-day-frame){
min-block-size:240px
}

.project-calendar :deep(.fc-dayGridWeek-view .calendar-event-title),.project-calendar :deep(.fc-dayGridDay-view .calendar-event-title){
white-space:normal;
display:-webkit-box;
-webkit-line-clamp:3;
-webkit-box-orient:vertical;
overflow:hidden
}

.project-calendar :deep(.fc-daygrid-day-number){
font-size:12px;
padding:7px 9px
}

.project-calendar :deep(.fc-day-today .fc-daygrid-day-number){
background:#2563eb;
color:white;
border-radius:50%;
margin:3px;
inline-size:28px;
block-size:28px;
box-sizing:border-box;
padding:0;
display:flex;
align-items:center;
justify-content:center;
flex-shrink:0
}

.project-calendar :deep(.fc-event){
border-radius:4px;
cursor:pointer
}

.project-calendar :deep(.fc-daygrid-event){
padding:2px 4px;
margin:2px 3px
}


.project-calendar :deep(.fc-daygrid-more-link){
font-size:12px;
color:#2563eb;
padding-inline:6px
}

.project-calendar :deep(.fc-popover){
background:var(--white);
color:var(--text);
border-radius:10px;
box-shadow:var(--shadow-md);
max-inline-size:430px
}

.project-calendar :deep(.fc-popover .calendar-event-title){
white-space:normal
}

.calendar-event-content{
min-inline-size:0;
overflow:hidden;
font-size:12px;
line-height:1.4
}


.calendar-event-title{
display:block;
white-space:nowrap;
overflow:hidden;
text-overflow:ellipsis;
word-break:normal;
overflow-wrap:normal
}

.project-calendar :deep(.calendar-done){
opacity:.5;
text-decoration:line-through
}

.calendar-hint{
color:var(--text-light);
font-size:12px;
margin:14px 0 0
}

.calendar-error{
padding:12px;
background:#fee2e2;
color:#991b1b;
margin-block-end:12px
}

@media(width <=800px){
.project-calendar{
padding:10px
}
.project-calendar :deep(.fc-toolbar){
flex-wrap:wrap;
gap:12px
}
.project-calendar :deep(.fc-toolbar-title){
font-size:18px
}
.project-calendar :deep(.fc-toolbar-chunk:nth-child(2)){
order:-1;
inline-size:100%
}

}

</style>
