<template>
	<div class="task-relations">
		<XButton
			v-if="editEnabled && Object.keys(relatedTasks).length > 0"
			id="showRelatedTasksFormButton"
			v-tooltip="$t('task.relation.add')"
			class="is-pulled-right add-task-relation-button d-print-none"
			:class="{'is-active': showNewRelationForm}"
			variant="secondary"
			icon="plus"
			:shadow="false"
			@click="showNewRelationForm = !showNewRelationForm"
		/>
		<transition-group name="fade">
			<template v-if="editEnabled && showCreate">
				<label
					key="label"
					class="label"
				>
					{{ $t('task.relation.new') }}
					<CustomTransition name="fade">
						<span
							v-if="taskRelationService.loading"
							class="is-inline-flex"
						>
							<span class="loader is-inline-block mie-2" />
							{{ $t('misc.saving') }}
						</span>
						<span
							v-else-if="!taskRelationService.loading && saved"
							class="has-text-success"
						>
							{{ $t('misc.saved') }}
						</span>
					</CustomTransition>
				</label>
				<div
					key="field-search"
					class="field"
				>
					<Multiselect
						v-model="newTaskRelation.task"
						v-focus
						:placeholder="$t('task.relation.searchPlaceholder')"
						:loading="taskService.loading"
						:search-results="mappedFoundTasks"
						label="title"
						:creatable="true"
						:create-placeholder="$t('task.relation.createPlaceholder')"
						@search="findTasks"
						@create="createAndRelateTask"
					>
						<template #searchResult="{option: task}">
							<span 
								v-if="typeof task !== 'string'"
								class="search-result"
								:class="{'is-strikethrough': task.done}"
							>
								<span
									v-if="task.projectId !== projectId"
									class="different-project"
								>
									<span
										v-if="task.differentProject !== null"
										v-tooltip="$t('task.relation.differentProject')"
									>
										{{ task.differentProject }} >
									</span>
								</span>
								{{ task.title }}
							</span>
							<span
								v-else
								class="search-result"
							>
								{{ task }}
							</span>
						</template>
					</Multiselect>
				</div>
				<div
					key="field-kind"
					class="field has-addons mbe-4"
				>
					<div class="control is-expanded">
						<div class="select is-fullwidth has-defaults">
							<select v-model="newTaskRelation.kind">
								<option value="unset">
									{{ $t('task.relation.select') }}
								</option>
								<option
									v-for="rk in RELATION_KINDS"
									:key="`option_${rk}`"
									:value="rk"
								>
									{{ $t(`task.relation.kinds.${rk}`, 1) }}
								</option>
							</select>
						</div>
					</div>
					<div class="control">
						<XButton @click="addTaskRelation()">
							{{ $t('task.relation.add') }}
						</XButton>
					</div>
				</div>
			</template>
		</transition-group>

		<div
			v-for="rts in mappedRelatedTasks"
			:key="rts.kind"
			class="related-tasks"
		>
			<span class="title">{{ rts.title }}</span>
			<div class="tasks">
				<div
					v-for="task in rts.tasks"
					:key="task.id"
					class="task"
				>
					<div class="is-flex is-align-items-center">
						<FancyCheckbox
							v-model="task.done"
							class="task-done-checkbox"
							@update:modelValue="toggleTaskDone(task)"
						/>
						<RouterLink
							:to="{ name: route.name as string, params: { id: task.id }, state: { backdropView: route.fullPath } }"
							:class="{ 'is-strikethrough': task.done}"
						>
							<span
								v-if="task.projectId !== projectId"
								class="different-project"
							>
								<span
									v-if="task.differentProject !== null"
									v-tooltip="$t('task.relation.differentProject')"
								>
									{{ task.differentProject }} >
								</span>
							</span>
							{{ task.title }}
						</RouterLink>
					</div>
					<BaseButton
						v-if="editEnabled"
						class="remove"
						@click="setRelationToDelete({
							relationKind: rts.kind,
							otherTaskId: task.id
						})"
					>
						<Icon icon="trash-alt" />
					</BaseButton>
				</div>
			</div>
		</div>
		<p
			v-if="showNoRelationsNotice && Object.keys(relatedTasks).length === 0"
			class="none"
		>
			{{ $t('task.relation.noneYet') }}
		</p>

		<Modal
			:enabled="relationToDelete !== undefined"
			@close="relationToDelete = undefined"
			@submit="removeTaskRelation()"
		>
			<template #header>
				<span>{{ $t('task.relation.delete') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('task.relation.deleteText1') }}<br>
					<strong class="has-text-white">{{ $t('misc.cannotBeUndone') }}</strong>
				</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {ref, reactive, shallowReactive, watch, computed} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRoute} from 'vue-router'

import TaskService from '@/services/task'
import TaskModel from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import type {ITaskRelation} from '@/modelTypes/ITaskRelation'
import {RELATION_KINDS, RELATION_KIND, type IRelationKind} from '@/types/IRelationKind'

import TaskRelationService from '@/services/taskRelation'
import TaskRelationModel from '@/models/taskRelation'

import CustomTransition from '@/components/misc/CustomTransition.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'

import {error, success} from '@/message'
import {useTaskStore} from '@/stores/tasks'
import {useProjectStore} from '@/stores/projects'
import {playPopSound} from '@/helpers/playPop'

const props = withDefaults(defineProps<{
	taskId: number,
	initialRelatedTasks?: ITask['relatedTasks'],
	showNoRelationsNotice?: boolean,
	projectId: number,
	editEnabled: boolean,
}>(), {
	initialRelatedTasks: () => ({}),
	showNoRelationsNotice: false,
})

const taskStore = useTaskStore()
const projectStore = useProjectStore()
const route = useRoute()
const {t} = useI18n({useScope: 'global'})

type TaskRelation = {kind: IRelationKind, task: ITask}

const taskService = shallowReactive(new TaskService())

const relatedTasks = ref<ITask['relatedTasks']>({})

const newTaskRelation: TaskRelation = reactive({
	kind: RELATION_KIND.RELATED,
	task: new TaskModel(),
})

watch(
	() => props.initialRelatedTasks,
	(value) => {
		relatedTasks.value = value
	},
	{immediate: true},
)

const showNewRelationForm = ref(false)
const showCreate = computed(() => Object.keys(relatedTasks.value).length === 0 || showNewRelationForm.value)

const query = ref('')
const foundTasks = ref<ITask[]>([])

async function findTasks(newQuery: string) {
	query.value = newQuery
	const result = await taskService.getAll({}, {
		s: newQuery,
		sort_by: 'done',
	})
	
	foundTasks.value = mapRelatedTasks(result)
}

function mapRelatedTasks(tasks: ITask[]) {
	return tasks.map(task => {
		// by doing this here once we can save a lot of duplicate calls in the template
		const project = projectStore.projects[task.projectId]

		return {
			...task,
			differentProject:
				(project &&
					task.projectId !== props.projectId &&
					project?.title) || null,
		}
	})
}

const mapRelationKindsTitleGetter = computed(() => ({
	'subtask': (count: number) => t('task.relation.kinds.subtask', count),
	'parenttask': (count: number) => t('task.relation.kinds.parenttask', count),
	'related': (count: number) => t('task.relation.kinds.related', count),
	'duplicateof': (count: number) => t('task.relation.kinds.duplicateof', count),
	'duplicates': (count: number) => t('task.relation.kinds.duplicates', count),
	'blocking': (count: number) => t('task.relation.kinds.blocking', count),
	'blocked': (count: number) => t('task.relation.kinds.blocked', count),
	'precedes': (count: number) => t('task.relation.kinds.precedes', count),
	'follows': (count: number) => t('task.relation.kinds.follows', count),
	'copiedfrom': (count: number) => t('task.relation.kinds.copiedfrom', count),
	'copiedto': (count: number) => t('task.relation.kinds.copiedto', count),
}))

const mappedRelatedTasks = computed(() => Object.entries(relatedTasks.value).map(
	([kind, tasks]) => ({
		title: mapRelationKindsTitleGetter.value[kind as IRelationKind](tasks.length),
		tasks: mapRelatedTasks(tasks),
		kind: kind as IRelationKind,
	}),
))
const mappedFoundTasks = computed(() => mapRelatedTasks(foundTasks.value.filter(t => t.id !== props.taskId)))

const taskRelationService = shallowReactive(new TaskRelationService())
const saved = ref(false)

async function addTaskRelation() {
	if (newTaskRelation.task.id === 0 && query.value !== '') {
		return createAndRelateTask(query.value)
	}

	if (newTaskRelation.task.id === 0) {
		error({message: t('task.relation.taskRequired')})
		return
	}

	await taskRelationService.create(new TaskRelationModel({
		taskId: props.taskId,
		otherTaskId: newTaskRelation.task.id,
		relationKind: newTaskRelation.kind,
	}))
	relatedTasks.value[newTaskRelation.kind] = [
		...(relatedTasks.value[newTaskRelation.kind] || []),
		newTaskRelation.task,
	]
	newTaskRelation.task = new TaskModel()
	saved.value = true
	showNewRelationForm.value = false
	setTimeout(() => {
		saved.value = false
	}, 2000)
}

const relationToDelete = ref<Partial<ITaskRelation>>()

function setRelationToDelete(relation: Partial<ITaskRelation>) {
	relationToDelete.value = relation
}

async function removeTaskRelation() {
	const relation = relationToDelete.value
	if (!relation || !relation.relationKind || !relation.otherTaskId) {
		relationToDelete.value = undefined
		return
	}
	try {
		const relationKind = relation.relationKind
		await taskRelationService.delete(new TaskRelationModel({
			relationKind,
			taskId: props.taskId,
			otherTaskId: relation.otherTaskId,
		}))

		relatedTasks.value[relationKind] = relatedTasks.value[relationKind]?.filter(
			({id}) => id !== relation.otherTaskId,
		)

		saved.value = true
		setTimeout(() => {
			saved.value = false
		}, 2000)
	} finally {
		relationToDelete.value = undefined
	}
}

async function createAndRelateTask(title: string) {
	const newTask = await taskService.create(new TaskModel({title, projectId: props.projectId}))
	newTaskRelation.task = newTask
	await addTaskRelation()
}

async function toggleTaskDone(task: ITask) {
	await taskStore.update(task)
	
	if (task.done) {
		playPopSound()
	}
	
	// Find the task in the project and update it so that it is correctly strike through
	Object.entries(relatedTasks.value).some(([kind, tasks]) => {
		return (tasks as ITask[]).some((t, key) => {
			const found = t.id === task.id
			if (found) {
				relatedTasks.value[kind as IRelationKind]![key] = task
			}
			return found
		})
	})

	success({message: t('task.detail.updateSuccess')})
}
</script>

<style lang="scss" scoped>
.add-task-relation-button {
	margin-block-start: -3rem;

	svg {
		transition: transform $transition;
	}

	&.is-active svg {
		transform: rotate(45deg);
	}
}

.different-project {
	color: var(--grey-500);
	inline-size: auto;
}

.title {
	font-size: 1rem;
	margin: 0;
}

.tasks {
	padding: .5rem;
}

.task {
	display: flex;
	flex-wrap: wrap;
	justify-content: space-between;
	padding: .75rem;
	transition: background-color $transition;
	border-radius: $radius;

	&:hover {
		background-color: var(--grey-200);
	}

	a {
		color: var(--text);
		transition: color ease $transition-duration;

		&:hover {
			color: var(--grey-900);
		}
	}

}

.remove {
	text-align: center;
	color: var(--danger);
	opacity: 0;
	transition: opacity $transition;
}

.task:hover .remove {
	opacity: 1;
}

.none {
	font-style: italic;
	text-align: center;
}

:deep(.multiselect .search-results button) {
	padding: 0.5rem;
}

// FIXME: The height of the actual checkbox in the <FancyCheckbox/> component is too much resulting in a 
//  weired positioning of the checkbox. Setting the height here is a workaround until we fix the styling 
//  of the component.
.task-done-checkbox {
	padding: 0;
	block-size: 18px; // The exact height of the checkbox in the container
	margin-inline-end: .75rem;
}
</style>
