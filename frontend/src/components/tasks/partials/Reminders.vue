<template>
	<div class="reminders">
		<div
			v-for="(r, index) in reminders"
			:key="index"
			:data-is-overdue="r.reminder && r.reminder < now || undefined"
			class="reminder-input"
		>
			<ReminderDetail
				v-model="reminders[index]"
				class="reminder-detail"
				:disabled="disabled"
				:default-relative-to="defaultRelativeTo"
				:allow-absolute="allowAbsolute"
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
			:allow-absolute="allowAbsolute"
			@update:modelValue="addNewReminder"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, watch} from 'vue'

import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import type {IReminderPeriodRelativeTo} from '@/types/IReminderPeriodRelativeTo'

import BaseButton from '@/components/base/BaseButton.vue'
import ReminderDetail from '@/components/tasks/partials/ReminderDetail.vue'
import {useNow} from '@vueuse/core'

const props = withDefaults(defineProps<{
	modelValue?: ITaskReminder[],
	defaultRelativeTo?: IReminderPeriodRelativeTo | null,
	disabled?: boolean,
	allowAbsolute?: boolean,
}>(), {
	modelValue: () => [],
	defaultRelativeTo: null,
	disabled: false,
	allowAbsolute: true,
})

const emit = defineEmits<{
	'update:modelValue': [ITaskReminder[]]
}>()

const reminders = ref<ITaskReminder[]>([])

const now = useNow({interval: 1000})

watch(
	() => props.modelValue,
	(newVal) => {
		reminders.value = [...(newVal ?? [])]
	},
	{immediate: true, deep: true}, // deep watcher so that we get the resolved date after updating the task
)

function updateData() {
	emit('update:modelValue', [...reminders.value])
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

	&[data-is-overdue] :deep(.datepicker .show) {
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
