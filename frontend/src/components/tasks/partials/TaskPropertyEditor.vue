<template>
	<template v-if="field === 'assignees'">
		<EditAssignees
			v-if="canWrite"
			:model-value="task.assignees"
			:project-id="task.projectId"
			:task-id="task.id"
			@update:modelValue="assignees => emit('mutate', {assignees})"
		/>
		<AssigneeList
			v-else-if="task.assignees.length > 0"
			:assignees="task.assignees"
		/>
		<span
			v-else
			class="property-empty"
		>{{ $t('misc.notSet') }}</span>
	</template>

	<template v-else-if="field === 'labels'">
		<EditLabels
			v-if="canWrite || task.labels.length > 0"
			:model-value="task.labels"
			:disabled="!canWrite"
			:task-id="task.id"
			:creatable="!authStore.isLinkShareAuth"
			:creation-disabled-message="authStore.isLinkShareAuth ? $t('task.label.linkShareCannotCreate') : ''"
			@update:modelValue="labels => emit('mutate', {labels})"
		/>
		<span
			v-else
			class="property-empty"
		>{{ $t('misc.notSet') }}</span>
	</template>

	<PrioritySelect
		v-else-if="field === 'priority'"
		:model-value="task.priority"
		:disabled="!canWrite"
		@update:modelValue="priority => emit('patch', {priority: priority as Priority})"
	/>

	<PercentDoneSelect
		v-else-if="field === 'percentDone'"
		:model-value="task.percentDone"
		:disabled="!canWrite"
		@update:modelValue="percentDone => emit('patch', {percentDone})"
	/>

	<Datepicker
		v-else-if="field === 'dueDate' || field === 'startDate' || field === 'endDate'"
		:model-value="task[field]"
		:choose-date-label="$t(DATE_LABELS[field])"
		:empty-label="$t('misc.notSet')"
		:disabled="loading || !canWrite"
		@update:modelValue="date => emit('mutate', {[field]: date})"
		@closeOnChange="emit('patch', {})"
	/>

	<Reminders
		v-else-if="field === 'reminders'"
		:model-value="task.reminders"
		:default-relative-to="remindersDefaultRelativeTo"
		:disabled="!canWrite"
		@update:modelValue="reminders => emit('patch', {reminders})"
	/>

	<RepeatAfter
		v-else-if="field === 'repeatAfter'"
		:model-value="task"
		:disabled="!canWrite"
		@update:modelValue="updated => updated && emit('patch', updated)"
	/>

	<template v-else-if="field === 'color'">
		<ColorPicker
			v-if="canWrite"
			:model-value="taskColor"
			menu-position="bottom"
			@update:modelValue="color => emit('update:taskColor', color)"
		/>
		<ColorBubble
			v-else-if="taskColor"
			:color="taskColor"
		/>
		<span
			v-else
			class="property-empty"
		>{{ $t('misc.notSet') }}</span>
	</template>
</template>

<script setup lang="ts">
import type {ITask} from '@/modelTypes/ITask'
import type {Priority} from '@/constants/priorities'
import type {IReminderPeriodRelativeTo} from '@/types/IReminderPeriodRelativeTo'

import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import ColorBubble from '@/components/misc/ColorBubble.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'
import Datepicker from '@/components/input/Datepicker.vue'
import EditAssignees from '@/components/tasks/partials/EditAssignees.vue'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'
import PercentDoneSelect from '@/components/tasks/partials/PercentDoneSelect.vue'
import PrioritySelect from '@/components/tasks/partials/PrioritySelect.vue'
import Reminders from '@/components/tasks/partials/Reminders.vue'
import RepeatAfter from '@/components/tasks/partials/RepeatAfter.vue'

import {useAuthStore} from '@/stores/auth'

export type TaskPropertyField =
	| 'assignees'
	| 'labels'
	| 'priority'
	| 'percentDone'
	| 'dueDate'
	| 'startDate'
	| 'endDate'
	| 'reminders'
	| 'repeatAfter'
	| 'color'

withDefaults(defineProps<{
	field: TaskPropertyField
	task: ITask
	canWrite: boolean
	taskColor: string
	remindersDefaultRelativeTo?: IReminderPeriodRelativeTo | null
	loading?: boolean
}>(), {
	remindersDefaultRelativeTo: null,
	loading: false,
})

const emit = defineEmits<{
	// Change the local task without saving (the editor saves itself or saves on close).
	'mutate': [patch: Partial<ITask>]
	// Apply the patch to the task and save it.
	'patch': [patch: Partial<ITask>]
	'update:taskColor': [color: string]
}>()

const DATE_LABELS = {
	dueDate: 'task.detail.chooseDueDate',
	startDate: 'task.detail.chooseStartDate',
	endDate: 'task.detail.chooseEndDate',
} as const

const authStore = useAuthStore()
</script>
