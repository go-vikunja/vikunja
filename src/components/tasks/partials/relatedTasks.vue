<template>
	<div class="task-relations">
		<x-button
			v-if="editEnabled && Object.keys(relatedTasks).length > 0"
			@click="showNewRelationForm = !showNewRelationForm"
			class="is-pulled-right add-task-relation-button d-print-none"
			:class="{'is-active': showNewRelationForm}"
			v-tooltip="$t('task.relation.add')"
			variant="secondary"
			icon="plus"
			:shadow="false"
		/>
		<transition-group name="fade">
			<template v-if="editEnabled && showCreate">
				<label class="label" key="label">
					{{ $t('task.relation.new') }}
					<CustomTransition name="fade">
						<span class="is-inline-flex" v-if="taskRelationService.loading">
							<span class="loader is-inline-block mr-2"></span>
							{{ $t('misc.saving') }}
						</span>
						<span class="has-text-success" v-else-if="!taskRelationService.loading && saved">
							{{ $t('misc.saved') }}
						</span>
					</CustomTransition>
				</label>
				<div class="field" key="field-search">
					<Multiselect
						:placeholder="$t('task.relation.searchPlaceholder')"
						@search="findTasks"
						:loading="taskService.loading"
						:search-results="mappedFoundTasks"
						label="title"
						v-model="newTaskRelation.task"
						:creatable="true"
						:create-placeholder="$t('task.relation.createPlaceholder')"
						@create="createAndRelateTask"
					>
						<template #searchResult="{option: task}">
							<span 
								v-if="typeof task !== 'string'"
								class="search-result"
								:class="{'is-strikethrough': task.done}"
							>
								<span
									class="different-list"
									v-if="task.listId !== listId"
								>
									<span
										v-if="task.differentNamespace !== null"
										v-tooltip="$t('task.relation.differentNamespace')">
										{{ task.differentNamespace }} >
									</span>
									<span
										v-if="task.differentList !== null"
										v-tooltip="$t('task.relation.differentList')">
										{{ task.differentList }} >
									</span>
								</span>
								{{ task.title }}
							</span>
							<span class="search-result" v-else>
								{{ task }}
							</span>
						</template>
					</Multiselect>
				</div>
				<div class="field has-addons mb-4" key="field-kind">
					<div class="control is-expanded">
						<div class="select is-fullwidth has-defaults">
							<select v-model="newTaskRelation.kind">
								<option value="unset">{{ $t('task.relation.select') }}</option>
								<option :key="`option_${rk}`" :value="rk" v-for="rk in RELATION_KINDS">
									{{ $tc(`task.relation.kinds.${rk}`, 1) }}
								</option>
							</select>
						</div>
					</div>
					<div class="control">
						<x-button @click="addTaskRelation()">{{ $t('task.relation.add') }}</x-button>
					</div>
				</div>
			</template>
		</transition-group>

		<div :key="rts.kind" class="related-tasks" v-for="rts in mappedRelatedTasks">
			<span class="title">{{ rts.title }}</span>
			<div class="tasks">
				<div :key="t.id" class="task" v-for="t in rts.tasks">
					<div class="is-flex is-align-items-center">
						<Fancycheckbox
							class="task-done-checkbox"
							v-model="t.done"
							@update:model-value="toggleTaskDone(t)"
						/>
						<router-link
							:to="{ name: route.name as string, params: { id: t.id } }"
							:class="{ 'is-strikethrough': t.done}"
						>
							<span
								class="different-list"
								v-if="t.listId !== listId"
							>
								<span
									v-if="t.differentNamespace !== null"
									v-tooltip="$t('task.relation.differentNamespace')">
									{{ t.differentNamespace }} >
								</span>
								<span
									v-if="t.differentList !== null"
									v-tooltip="$t('task.relation.differentList')">
									{{ t.differentList }} >
								</span>
							</span>
							{{ t.title }}
						</router-link>
					</div>
					<BaseButton
						v-if="editEnabled"
						@click="setRelationToDelete({
							relationKind: rts.kind,
							otherTaskId: t.id
						})"
						class="remove"
					>
						<icon icon="trash-alt"/>
					</BaseButton>
				</div>
			</div>
		</div>
		<p class="none" v-if="showNoRelationsNotice && Object.keys(relatedTasks).length === 0">
			{{ $t('task.relation.noneYet') }}
		</p>

		<modal
			:enabled="relationToDelete !== undefined"
			@close="relationToDelete = undefined"
			@submit="removeTaskRelation()"
		>
			<template #header><span>{{ $t('task.relation.delete') }}</span></template>

			<template #text>
				<p>
					{{ $t('task.relation.deleteText1') }}<br/>
					<strong class="has-text-white">{{ $t('misc.cannotBeUndone') }}</strong>
				</p>
			</template>
		</modal>
	</div>
</template>

<script setup lang="ts">
import {ref, reactive, shallowReactive, watch, computed, type PropType} from 'vue'
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
import Multiselect from '@/components/input/multiselect.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'

import {useNamespaceStore} from '@/stores/namespaces'

import {error, success} from '@/message'
import {useTaskStore} from '@/stores/tasks'

const props = defineProps({
	taskId: {
		type: Number,
		required: true,
	},
	initialRelatedTasks: {
		type: Object as PropType<ITask['relatedTasks']>,
		default: () => ({}),
	},
	showNoRelationsNotice: {
		type: Boolean,
		default: false,
	},
	listId: {
		type: Number,
		default: 0,
	},
	editEnabled: {
		default: true,
	},
})

const taskStore = useTaskStore()
const namespaceStore = useNamespaceStore()
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
	foundTasks.value = await taskService.getAll({}, {s: newQuery})
}

const getListAndNamespaceById = (listId: number) => namespaceStore.getListAndNamespaceById(listId, true)

const namespace = computed(() => getListAndNamespaceById(props.listId)?.namespace)

function mapRelatedTasks(tasks: ITask[]) {
	return tasks.map(task => {
		// by doing this here once we can save a lot of duplicate calls in the template
		const {
			list,
			namespace: taskNamespace,
		} = getListAndNamespaceById(task.listId) || {list: null, namespace: null}

		return {
			...task,
			differentNamespace:
				(taskNamespace !== null &&
					taskNamespace.id !== namespace.value.id &&
					taskNamespace?.title) || null,
			differentList:
				(list !== null &&
					task.listId !== props.listId &&
					list?.title) || null,
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
	const newTask = await taskService.create(new TaskModel({title, listId: props.listId}))
	newTaskRelation.task = newTask
	await addTaskRelation()
}

async function toggleTaskDone(task: ITask) {
	await taskStore.update(task)
	
	// Find the task in the list and update it so that it is correctly strike through
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
	margin-top: -3rem;

	svg {
		transition: transform $transition;
	}

	&.is-active svg {
		transform: rotate(45deg);
	}
}

.different-list {
	color: var(--grey-500);
	width: auto;
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

// FIXME: The height of the actual checkbox in the <Fancycheckbox/> component is too much resulting in a 
//  weired positioning of the checkbox. Setting the height here is a workaround until we fix the styling 
//  of the component.
.task-done-checkbox {
	padding: 0;
	height: 18px; // The exact height of the checkbox in the container
}
</style>