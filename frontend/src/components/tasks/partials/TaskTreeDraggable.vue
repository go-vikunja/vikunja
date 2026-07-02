<template>
	<draggable
		:model-value="tasks"
		:group="{name: 'tasks', put: !disabled}"
		:disabled="disabled"
		:sort="!disabled"
		:item-key="(task: ITask) => task.id"
		tag="ul"
		:component-data="{
			class: {
				tasks: true,
				'task-tree': true,
				'task-tree--nested': nested,
				'task-tree--empty': dragging && tasks.length === 0,
				'dragging-disabled': disabled,
			},
			'data-parent-task-id': parentTaskId ?? '',
			type: transitionGroup ? 'transition-group' : undefined,
		}"
		:animation="100"
		:handle="handle"
		:delay-on-touch-only="delayOnTouchOnly"
		:delay="delay"
		ghost-class="task-ghost"
		:move="canMoveTask"
		@update:modelValue="updateList"
		@start="$emit('dragStart', $event)"
		@end="emitDrop"
	>
		<template #item="{element: task, index}">
			<div
				class="task-tree-item"
				:data-task-id="task.id"
			>
				<SingleTaskInProject
					:ref="(el) => taskRefSetter?.(el, index)"
					:show-list-color="showListColor"
					:show-project="showProject && !nested"
					:can-mark-as-done="canMarkTaskAsDone(task)"
					:the-task="task"
					:all-tasks="[]"
					@taskUpdated="$emit('taskUpdated', $event)"
				>
					<span
						v-if="showDragHandle && !disabled"
						class="icon handle"
					>
						<Icon icon="grip-lines" />
					</span>
				</SingleTaskInProject>

				<TaskTreeDraggable
					v-if="hierarchical"
					:tasks="getSubtasks(task)"
					:all-tasks="allTasks"
					:hierarchical="hierarchical"
					:parent-task-id="task.id"
					:nested="true"
					:disabled="disabled"
					:show-project="showProject"
					:show-list-color="showListColor"
					:show-drag-handle="showDragHandle"
					:dragging="dragging"
					:handle="handle"
					:delay="delay"
					:delay-on-touch-only="delayOnTouchOnly"
					:can-mark-task-as-done="canMarkTaskAsDone"
					@taskUpdated="$emit('taskUpdated', $event)"
					@dragStart="$emit('dragStart', $event)"
					@drop="$emit('drop', $event)"
					@updateList="$emit('updateList', $event)"
				/>
			</div>
		</template>
	</draggable>
</template>

<script setup lang="ts">
import draggable from 'zhyswan-vuedraggable'

import Icon from '@/components/misc/Icon'
import SingleTaskInProject from '@/components/tasks/partials/SingleTaskInProject.vue'
import type {ITask} from '@/modelTypes/ITask'

const props = withDefaults(defineProps<{
	tasks: ITask[],
	allTasks: ITask[],
	hierarchical: boolean,
	parentTaskId?: ITask['id'] | null,
	nested?: boolean,
	disabled?: boolean,
	showProject?: boolean,
	showListColor?: boolean,
	showDragHandle?: boolean,
	dragging?: boolean,
	handle?: string,
	delay?: number,
	delayOnTouchOnly?: boolean,
	transitionGroup?: boolean,
	canMarkTaskAsDone: (task: ITask) => boolean,
	taskRefSetter?: (el: unknown, index: number) => void,
}>(), {
	parentTaskId: null,
	nested: false,
	disabled: false,
	showProject: false,
	showListColor: false,
	showDragHandle: false,
	dragging: false,
	handle: undefined,
	delay: 0,
	delayOnTouchOnly: false,
	transitionGroup: false,
	taskRefSetter: undefined,
})

const emit = defineEmits<{
	taskUpdated: [task: ITask],
	dragStart: [event: { item: HTMLElement }],
	drop: [event: TaskTreeDropEvent],
	updateList: [event: TaskTreeListUpdateEvent],
}>()

defineOptions({name: 'TaskTreeDraggable'})

export interface TaskTreeDropEvent {
	taskId: ITask['id']
	oldParentTaskId: ITask['id'] | null
	newParentTaskId: ITask['id'] | null
	oldIndex: number
	newIndex: number
	originalEvent?: MouseEvent
	from: HTMLElement
	to: HTMLElement
}

export interface TaskTreeListUpdateEvent {
	parentTaskId: ITask['id'] | null
	tasks: ITask[]
}

function parseParentTaskId(element: HTMLElement): ITask['id'] | null {
	const value = element.dataset.parentTaskId
	if (!value) {
		return null
	}

	const parsed = parseInt(value, 10)
	return Number.isNaN(parsed) ? null : parsed
}

function getSubtasks(task: ITask): ITask[] {
	if (!props.hierarchical) {
		return []
	}

	return (task.relatedTasks?.subtask ?? [])
		.map(subtask => props.allTasks.find(t => t.id === subtask.id) ?? subtask)
}

function isDescendant(task: ITask, possibleDescendantId: ITask['id']): boolean {
	return getSubtasks(task).some(subtask => (
		subtask.id === possibleDescendantId || isDescendant(subtask, possibleDescendantId)
	))
}

function canMoveTask(event: { draggedContext?: { element?: ITask }, to: HTMLElement }) {
	const draggedTask = event.draggedContext?.element
	if (!draggedTask) {
		return true
	}

	const targetParentId = parseParentTaskId(event.to)
	return targetParentId === null ||
		(targetParentId !== draggedTask.id && !isDescendant(draggedTask, targetParentId))
}

function updateList(updatedTasks: ITask[]) {
	emit('updateList', {
		parentTaskId: props.parentTaskId,
		tasks: updatedTasks,
	})
}

function emitDrop(event: {
	item: HTMLElement,
	from: HTMLElement,
	to: HTMLElement,
	oldIndex: number,
	newIndex: number,
	originalEvent?: MouseEvent,
}) {
	const taskId = parseInt(event.item.dataset.taskId ?? '', 10)
	if (Number.isNaN(taskId)) {
		return
	}

	emit('drop', {
		taskId,
		oldParentTaskId: parseParentTaskId(event.from),
		newParentTaskId: parseParentTaskId(event.to),
		oldIndex: event.oldIndex,
		newIndex: event.newIndex,
		originalEvent: event.originalEvent,
		from: event.from,
		to: event.to,
	})
}
</script>

<style lang="scss" scoped>
.task-tree {
	padding: .5rem;
}

.task-tree--nested {
	margin-inline-start: 1.75rem;
	padding-block: 0;
}

.task-tree--empty {
	min-block-size: .5rem;
	padding-block: .25rem;
}

.task-ghost {
	border-radius: $radius;
	background: var(--grey-100);
	border: 2px dashed var(--grey-300);

	* {
		opacity: 0;
	}
}
</style>
