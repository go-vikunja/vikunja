<template>
	<div class="reminders">
		<div
			v-for="(r, index) in reminders"
			:key="index"
			:class="{ 'overdue': r.reminder < now }"
			class="reminder-input"
		>
			<ReminderDetail
				v-model="reminders[index]"
				class="reminder-detail"
				:disabled="disabled"
				:default-relative-to="defaultRelativeTo"
				@update:modelValue="updateData"
			/>
			<BaseButton
				v-if="!disabled"
				class="remove"
				@click="removeReminderByIndex(index)"
			>
				<Icon icon="times" />
			</BaseButton>
		</div>

		<ReminderDetail
			:disabled="disabled"
			:clear-after-update="true"
			:default-relative-to="defaultRelativeTo"
			@update:modelValue="addNewReminder"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, watch, computed} from 'vue'

import type {ITaskReminder} from '@/modelTypes/ITaskReminder'

import BaseButton from '@/components/base/BaseButton.vue'
import ReminderDetail from '@/components/tasks/partials/ReminderDetail.vue'
import type {ITask} from '@/modelTypes/ITask'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'
import { useNow } from '@vueuse/core'

const props = withDefaults(defineProps<{
	modelValue: ITask,
	disabled?: boolean,
}>(), {
	disabled: false,
})

const emit = defineEmits<{
	'update:modelValue': [ITask]
}>()

const reminders = ref<ITaskReminder[]>([])

const now = useNow({interval: 1000})

watch(
	() => props.modelValue.reminders,
	(newVal) => {
		reminders.value = newVal
	},
	{immediate: true, deep: true}, // deep watcher so that we get the resolved date after updating the task
)

const defaultRelativeTo = computed(() => {
	if (typeof props.modelValue === 'undefined') {
		return null
	}
	
	if (props.modelValue?.dueDate) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE
	}
	
	if (props.modelValue.dueDate === null && props.modelValue.startDate !== null) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE
	}
	
	if (props.modelValue.dueDate === null && props.modelValue.startDate === null && props.modelValue.endDate !== null) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE
	}
	
	return null
})

function updateData() {
	emit('update:modelValue', {
		...props.modelValue,
		reminders: reminders.value,
	})
}

function addNewReminder(newReminder: ITaskReminder|null) {
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

	&:last-child {
		margin-block-end: 0.75rem;
	}
}

.reminder-detail {
	inline-size: 100%;
}

.remove {
	color: var(--danger);
	vertical-align: top;
	padding-inline-start: .5rem;
	line-height: 1;
}
</style>
