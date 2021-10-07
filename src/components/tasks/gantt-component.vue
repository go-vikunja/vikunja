<template>
	<div class="gantt-chart">
		<div class="filter-container">
			<div class="items">
				<x-button
					@click.prevent.stop="showTaskFilter = !showTaskFilter"
					type="secondary"
					icon="filter"
				>
					{{ $t('filters.title') }}
				</x-button>
			</div>
			<filter-popup
				:visible="showTaskFilter"
				v-model="params"
				@update:modelValue="loadTasks()"
			/>
		</div>
		<div class="dates">
			<template v-for="(y, yk) in days" :key="yk + 'year'">
				<div class="months">
					<div
						:key="mk + 'month'"
						class="month"
						v-for="(m, mk) in days[yk]"
					>
						{{ formatYear(new Date(`${yk}-${parseInt(mk) + 1}-01`)) }}
						<div class="days">
							<div
								:class="{ today: d.toDateString() === now.toDateString() }"
								:key="dk + 'day'"
								:style="{ width: dayWidth + 'px' }"
								class="day"
								v-for="(d, dk) in days[yk][mk]"
							>
								<span class="theday" v-if="dayWidth > 25">
									{{ d.getDate() }}
								</span>
								<span class="weekday" v-if="dayWidth > 25">
									{{
										d.toLocaleString('en-us', {
											weekday: 'short',
										})
									}}
								</span>
							</div>
						</div>
					</div>
				</div>
			</template>
		</div>
		<div :style="{ width: fullWidth + 'px' }" class="tasks">
			<div
				v-for="(t, k) in theTasks"
				:key="t ? t.id : 0"
				:style="{
					background:
						'repeating-linear-gradient(90deg, #ededed, #ededed 1px, ' +
						(k % 2 === 0
							? '#fafafa 1px, #fafafa '
							: '#fff 1px, #fff ') +
						dayWidth +
						'px)',
				}"
				class="row"
			>
				<VueDragResize
					:class="{
						done: t ? t.done : false,
						'is-current-edit': taskToEdit !== null && taskToEdit.id === t.id,
						'has-light-text': !colorIsDark(t.getHexColor()),
						'has-dark-text': colorIsDark(t.getHexColor()),
					}"
					:gridX="dayWidth"
					:h="31"
					:isActive="canWrite"
					:minw="dayWidth"
					:parentLimitation="true"
					:parentW="fullWidth"
					:snapToGrid="true"
					:sticks="['mr', 'ml']"
					:style="{
						'border-color': t.getHexColor(),
						'background-color': t.getHexColor(),
					}"
					:w="t.durationDays * dayWidth"
					:x="t.offsetDays * dayWidth - 6"
					:y="0"
					@dragstop="(e) => resizeTask(t, e)"
					@resizestop="(e) => resizeTask(t, e)"
					axis="x"
					class="task"
				>
					<span
						:class="{
							'has-high-priority': t.priority >= priorities.HIGH,
							'has-not-so-high-priority':
								t.priority === priorities.HIGH,
							'has-super-high-priority':
								t.priority === priorities.DO_NOW,
						}"
					>
						{{ t.title }}
					</span>
					<priority-label :priority="t.priority" :done="t.done"/>
					<!-- using the key here forces vue to use the updated version model and not the response returned by the api -->
					<a @click="editTask(theTasks[k])" class="edit-toggle">
						<icon icon="pen"/>
					</a>
				</VueDragResize>
			</div>
			<template v-if="showTaskswithoutDates">
				<div
					:key="t.id"
					:style="{
						background:
							'repeating-linear-gradient(90deg, #ededed, #ededed 1px, ' +
							(k % 2 === 0
								? '#fafafa 1px, #fafafa '
								: '#fff 1px, #fff ') +
							dayWidth +
							'px)',
					}"
					class="row"
					v-for="(t, k) in tasksWithoutDates"
				>
					<VueDragResize
						:gridX="dayWidth"
						:h="31"
						:isActive="canWrite"
						:minw="dayWidth"
						:parentLimitation="true"
						:parentW="fullWidth"
						:snapToGrid="true"
						:sticks="['mr', 'ml']"
						:x="dayOffsetUntilToday * dayWidth - 6"
						:y="0"
						@dragstop="(e) => resizeTask(t, e)"
						@resizestop="(e) => resizeTask(t, e)"
						axis="x"
						class="task nodate"
						v-tooltip="$t('list.gantt.noDates')"
					>
						<span>{{ t.title }}</span>
					</VueDragResize>
				</div>
			</template>
		</div>
		<form
			@submit.prevent="addNewTask()"
			class="add-new-task"
			v-if="canWrite"
		>
			<transition name="width">
				<input
					@blur="hideCrateNewTask"
					@keyup.esc="newTaskFieldActive = false"
					class="input"
					ref="newTaskTitleField"
					type="text"
					v-if="newTaskFieldActive"
					v-model="newTaskTitle"
				/>
			</transition>
			<x-button @click="showCreateNewTask" :shadow="false" icon="plus">
				{{ $t('list.list.newTaskCta') }}
			</x-button>
		</form>
		<transition name="fade">
			<card
				v-if="isTaskEdit"
				class="taskedit"
				:title="$t('list.list.editTask')"
				@close="() => {isTaskEdit = false;taskToEdit = null}"
				:has-close="true"
			>
				<edit-task :task="taskToEdit"/>
			</card>
		</transition>
	</div>
</template>

<script>
import VueDragResize from 'vue-drag-resize'
import EditTask from './edit-task'

import TaskService from '../../services/task'
import TaskModel from '../../models/task'
import priorities from '../../models/constants/priorities'
import PriorityLabel from './partials/priorityLabel'
import TaskCollectionService from '../../services/taskCollection'
import {mapState} from 'vuex'
import Rights from '../../models/constants/rights.json'
import FilterPopup from '@/components/list/partials/filter-popup.vue'

export default {
	name: 'GanttChart',
	components: {
		FilterPopup,
		PriorityLabel,
		EditTask,
		VueDragResize,
	},
	props: {
		listId: {
			type: Number,
			required: true,
		},
		showTaskswithoutDates: {
			type: Boolean,
			default: false,
		},
		dateFrom: {
			default: new Date(new Date().setDate(new Date().getDate() - 15)),
		},
		dateTo: {
			default: new Date(new Date().setDate(new Date().getDate() + 30)),
		},
		// The width of a day in pixels, used to calculate all sorts of things.
		dayWidth: {
			type: Number,
			default: 35,
		},
	},
	data() {
		return {
			days: [],
			startDate: null,
			endDate: null,
			theTasks: [], // Pretty much a copy of the prop, since we cant mutate the prop directly
			tasksWithoutDates: [],
			taskService: new TaskService(),
			fullWidth: 0,
			now: new Date(),
			dayOffsetUntilToday: 0,
			isTaskEdit: false,
			taskToEdit: null,
			newTaskTitle: '',
			newTaskFieldActive: false,
			priorities: priorities,
			taskCollectionService: new TaskCollectionService(),
			showTaskFilter: false,

			params: {
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter_by: ['done'],
				filter_value: ['false'],
				filter_comparator: ['equals'],
				filter_concat: 'and',
			},
		}
	},
	watch: {
		dateFrom: 'buildTheGanttChart',
		dateTo: 'buildTheGanttChart',
		listId: 'parseTasks',
	},
	mounted() {
		this.buildTheGanttChart()
	},
	computed: mapState({
		canWrite: (state) => state.currentList.maxRight > Rights.READ,
	}),
	methods: {
		buildTheGanttChart() {
			this.setDates()
			this.prepareGanttDays()
			this.parseTasks()
		},
		setDates() {
			this.startDate = new Date(this.dateFrom)
			this.endDate = new Date(this.dateTo)
			console.debug('setDates; start date: ', this.startDate, 'end date:', this.endDate, 'date from:', this.dateFrom, 'date to:', this.dateTo)

			this.dayOffsetUntilToday = Math.floor((this.now - this.startDate) / 1000 / 60 / 60 / 24) + 1
		},
		prepareGanttDays() {
			console.debug('prepareGanttDays; start date: ', this.startDate, 'end date:', this.endDate)
			// Layout: years => [months => [days]]
			let years = {}
			for (
				let d = this.startDate;
				d <= this.endDate;
				d.setDate(d.getDate() + 1)
			) {
				let date = new Date(d)
				if (years[date.getFullYear() + ''] === undefined) {
					years[date.getFullYear() + ''] = {}
				}
				if (years[date.getFullYear() + ''][date.getMonth() + ''] === undefined) {
					years[date.getFullYear() + ''][date.getMonth() + ''] = []
				}
				years[date.getFullYear() + ''][date.getMonth() + ''].push(date)
				this.fullWidth += this.dayWidth
			}
			console.debug('prepareGanttDays; years:', years)
			this.days = years
		},
		parseTasks() {
			this.setDates()
			this.loadTasks()
		},
		loadTasks() {
			this.theTasks = []
			this.tasksWithoutDates = []

			const getAllTasks = (page = 1) => {
				return this.taskCollectionService
					.getAll({listId: this.listId}, this.params, page)
					.then((tasks) => {
						if (page < this.taskCollectionService.totalPages) {
							return getAllTasks(page + 1).then((nextTasks) => {
								return tasks.concat(nextTasks)
							})
						} else {
							return tasks
						}
					})
					.catch((e) => {
						return Promise.reject(e)
					})
			}

			getAllTasks()
				.then((tasks) => {
					this.theTasks = tasks
						.filter((t) => {
							if (t.startDate === null && !t.done) {
								this.tasksWithoutDates.push(t)
							}
							return (
								t.startDate >= this.startDate &&
								t.endDate <= this.endDate
							)
						})
						.map((t) => {
							return this.addGantAttributes(t)
						})
						.sort(function (a, b) {
							if (a.startDate < b.startDate) return -1
							if (a.startDate > b.startDate) return 1
							return 0
						})
				})
				.catch((e) => {
					this.$message.error(e)
				})
		},
		addGantAttributes(t) {
			if (typeof t.durationDays !== 'undefined' && typeof t.offsetDays !== 'undefined') {
				return t
			}

			t.endDate === null ? this.endDate : t.endDate
			t.durationDays = Math.floor((t.endDate - t.startDate) / 1000 / 60 / 60 / 24)
			t.offsetDays = Math.floor((t.startDate - this.startDate) / 1000 / 60 / 60 / 24)
			return t
		},
		resizeTask(taskDragged, newRect) {
			if (this.isTaskEdit) {
				return
			}

			let newTask = { ...taskDragged }

			const didntHaveDates = newTask.startDate === null ? true : false

			let startDate = new Date(this.startDate)
			startDate.setDate(
				startDate.getDate() + newRect.left / this.dayWidth,
			)
			startDate.setUTCHours(0)
			startDate.setUTCMinutes(0)
			startDate.setUTCSeconds(0)
			startDate.setUTCMilliseconds(0)
			newTask.startDate = startDate
			let endDate = new Date(startDate)
			endDate.setDate(
				startDate.getDate() + newRect.width / this.dayWidth,
			)
			newTask.startDate = startDate
			newTask.endDate = endDate

			// We take the task from the overall tasks array because the one in it has bad data after it was updated once.
			// FIXME: This is a workaround. We should use a better mechanism to get the task or, even better,
			// prevent it from containing outdated Data in the first place.
			for (const tt in this.theTasks) {
				if (this.theTasks[tt].id === newTask.id) {
					newTask = this.theTasks[tt]
					break
				}
			}

			const ganttData = {
				endDate: newTask.endDate,
				durationDays: newTask.durationDays,
				offsetDays: newTask.offsetDays,
			}

			this.taskService
				.update(newTask)
				.then(r => {
					r.endDate = ganttData.endDate
					r.durationDays = ganttData.durationDays
					r.offsetDays = ganttData.offsetDays

					// If the task didn't have dates before, we'll update the list
					if (didntHaveDates) {
						for (const t in this.tasksWithoutDates) {
							if (this.tasksWithoutDates[t].id === r.id) {
								this.tasksWithoutDates.splice(t, 1)
								break
							}
						}
						this.theTasks.push(this.addGantAttributes(r))
					} else {
						for (const tt in this.theTasks) {
							if (this.theTasks[tt].id === r.id) {
								this.theTasks[tt] = this.addGantAttributes(r)
								break
							}
						}
					}
				})
				.catch((e) => {
					this.$message.error(e)
				})
		},
		editTask(task) {
			this.taskToEdit = task
			this.isTaskEdit = true
		},
		showCreateNewTask() {
			if (!this.newTaskFieldActive) {
				// Timeout to not send the form if the field isn't even shown
				setTimeout(() => {
					this.newTaskFieldActive = true
					this.$nextTick(() => this.$refs.newTaskTitleField.focus())
				}, 100)
			}
		},
		hideCrateNewTask() {
			if (this.newTaskTitle === '') {
				this.$nextTick(() => (this.newTaskFieldActive = false))
			}
		},
		addNewTask() {
			if (!this.newTaskFieldActive) {
				return
			}
			let task = new TaskModel({
				title: this.newTaskTitle,
				listId: this.listId,
			})
			this.taskService
				.create(task)
				.then((r) => {
					this.tasksWithoutDates.push(this.addGantAttributes(r))
					this.newTaskTitle = ''
					this.hideCrateNewTask()
				})
				.catch((e) => {
					this.$message.error(e)
				})
		},
		formatYear(date) {
			return this.format(date, 'MMMM, yyyy')
		},
	},
}
</script>
