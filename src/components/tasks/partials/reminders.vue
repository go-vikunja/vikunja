<template>
	<div class="reminders">
		<div
			v-for="(r, index) in reminders"
			:key="index"
			:class="{ 'overdue': r.reminder < new Date() }"
			class="reminder-input"
		>
			<ReminderDetail
				class="reminder-detail"
				:disabled="disabled"
				v-model="reminders[index]"
				@update:model-value="updateData"
				:default-relative-to="defaultRelativeTo"
			/>
			<BaseButton
				v-if="!disabled"
				@click="removeReminderByIndex(index)"
				class="remove"
			>
				<icon icon="times"/>
			</BaseButton>
		</div>

		<ReminderDetail
			:disabled="disabled"
			@update:modelValue="addNewReminder"
			:clear-after-update="true"
			:default-relative-to="defaultRelativeTo"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, watch, computed} from 'vue'

import type {ITaskReminder} from '@/modelTypes/ITaskReminder'

import BaseButton from '@/components/base/BaseButton.vue'
import ReminderDetail from '@/components/tasks/partials/reminder-detail.vue'
import type {ITask} from '@/modelTypes/ITask'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'

const {
	modelValue,
	disabled = false,
} = defineProps<{
	modelValue: ITask,
	disabled?: boolean,
}>()

const emit = defineEmits(['update:modelValue'])

const reminders = ref<ITaskReminder[]>([])

watch(
	() => modelValue.reminders,
	(newVal) => {
		reminders.value = newVal
	},
	{immediate: true, deep: true}, // deep watcher so that we get the resolved date after updating the task
)

const defaultRelativeTo = computed(() => {
	if (typeof modelValue === 'undefined') {
		return null
	}
	
	if (modelValue?.dueDate) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE
	}
	
	if (modelValue.dueDate === null && modelValue.startDate !== null) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE
	}
	
	if (modelValue.dueDate === null && modelValue.startDate === null && modelValue.endDate !== null) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE
	}
	
	return null
})

function updateData() {
	emit('update:modelValue', {
		...modelValue,
		reminders: reminders.value,
	})
}

function addNewReminder(newReminder: ITaskReminder) {
	if (newReminder === null) {
		return
	}
	reminders.value.push(newReminder)
	updateData()
}

function removeReminderByIndex(index: number) {
	reminders.value.splice(index, 1)
	updateData()
}
</script>

<style lang="scss" scoped>
.reminder-input {
	display: flex;
	align-items: center;

	&.overdue :deep(.datepicker .show) {
		color: var(--danger);
	}

	&::last-child {
		margin-bottom: 0.75rem;
	}
}

.reminder-detail {
	width: 100%;
}

.remove {
	color: var(--danger);
	vertical-align: top;
	padding-left: .5rem;
	line-height: 1;
}
</style>