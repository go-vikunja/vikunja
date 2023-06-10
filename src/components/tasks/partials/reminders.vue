<template>
	<div class="reminders">
		<div
			v-for="(r, index) in reminders"
			:key="index"
			:class="{ 'overdue': r.reminder < new Date() }"
			class="reminder-input"
		>
			<div class="reminder-detail">
				<ReminderDetail
					:disabled="disabled"
					v-model="reminders[index]"
					@update:model-value="updateData"/>
			</div>
			<div>
				<BaseButton v-if="!disabled" @click="removeReminderByIndex(index)" class="remove">
					<icon icon="times"/>
				</BaseButton>
			</div>
		</div>

		<ReminderDetail
			:disabled="disabled"
			@update:modelValue="addNewReminder"
			:clear-after-update="true"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, watch, type PropType} from 'vue'

import type {ITaskReminder} from '@/modelTypes/ITaskReminder'

import BaseButton from '@/components/base/BaseButton.vue'
import ReminderDetail from '@/components/tasks/partials/reminder-detail.vue'

const props = defineProps({
	modelValue: {
		type: Array as PropType<ITaskReminder[]>,
		default: () => [],
	},
	disabled: {
		default: false,
	},
})

const emit = defineEmits(['update:modelValue'])

const reminders = ref<ITaskReminder[]>([])

watch(
	props.modelValue,
	(newVal) => {
		reminders.value = newVal
	},
	{immediate: true},
)

function updateData() {
	emit('update:modelValue', reminders.value)
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