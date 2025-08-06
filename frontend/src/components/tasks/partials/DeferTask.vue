<template>
	<div
		:class="{ 'is-loading': taskService.loading }"
		class="defer-task loading-container"
	>
		<label class="label">{{ $t('task.deferDueDate.title') }}</label>
		<div class="defer-days">
			<XButton
				:shadow="false"
				variant="secondary"
				@click.prevent.stop="() => deferDays(1)"
			>
				{{ $t('task.deferDueDate.1day') }}
			</XButton>
			<XButton
				:shadow="false"
				variant="secondary"
				@click.prevent.stop="() => deferDays(3)"
			>
				{{ $t('task.deferDueDate.3days') }}
			</XButton>
			<XButton
				:shadow="false"
				variant="secondary"
				@click.prevent.stop="() => deferDays(7)"
			>
				{{ $t('task.deferDueDate.1week') }}
			</XButton>
		</div>
		<flat-pickr
			v-model="dueDate"
			:class="{ disabled: taskService.loading }"
			:config="flatPickerConfig"
			:disabled="taskService.loading || undefined"
			class="input"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, shallowReactive, computed, watch, onMounted, onBeforeUnmount} from 'vue'
import {useI18n} from 'vue-i18n'
import flatPickr from 'vue-flatpickr-component'

import TaskService from '@/services/task'
import type {ITask} from '@/modelTypes/ITask'
import {useFlatpickrLanguage} from '@/helpers/useFlatpickrLanguage'

const props = defineProps<{
	modelValue: ITask,
}>()

const emit = defineEmits<{
	'update:modelValue': [value: ITask]
}>()

const {t} = useI18n({useScope: 'global'})

const taskService = shallowReactive(new TaskService())
const task = ref<ITask>()

// We're saving the due date separately to prevent null errors in very short periods where the task is null.
const dueDate = ref<Date | null>()
const lastValue = ref<Date | null>()
const changeInterval = ref<ReturnType<typeof setInterval>>()

watch(
	() => props.modelValue,
	(value) => {
		task.value = { ...value }
		dueDate.value = value.dueDate
		lastValue.value = value.dueDate
	},
	{immediate: true},
)

onMounted(() => {
	// Because we don't really have other ways of handling change since if we let flatpickr
	// change events trigger updates, it would trigger a flatpickr change event which would trigger
	// an update which would trigger a change event and so on...
	// This is either a bug in flatpickr or in the vue component of it.
	// To work around that, we're only updating if something changed and check each second and when closing the popup.
	if (changeInterval.value) {
		clearInterval(changeInterval.value)
	}

	changeInterval.value = setInterval(updateDueDate, 1000)
})

onBeforeUnmount(() => {
	if (changeInterval.value) {
		clearInterval(changeInterval.value)
	}
	updateDueDate()
})

const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: true,
	time_24hr: true,
	inline: true,
	locale: useFlatpickrLanguage().value,
}))

function deferDays(days: number) {
	dueDate.value = new Date(dueDate.value)
	const currentDate = new Date(dueDate.value).getDate()
	dueDate.value = new Date(dueDate.value).setDate(currentDate + days)
	updateDueDate()
}

async function updateDueDate() {
	if (!dueDate.value) {
		return
	}

	if (+new Date(dueDate.value) === +lastValue.value) {
		return
	}

	const newTask = await taskService.update({
		...task.value,
		dueDate: new Date(dueDate.value),
	})
	lastValue.value = newTask.dueDate
	task.value = newTask
	emit('update:modelValue', newTask)
}
</script>

<style lang="scss" scoped>
// 100px is roughly the size the pane is pulled to the right
$defer-task-max-width: 350px + 100px;

.defer-task {
	inline-size: 100%;
	max-inline-size: $defer-task-max-width;

	@media screen and (max-width: ($defer-task-max-width)) {
		inset-inline-start: .5rem;
		inset-inline-end: .5rem;
		max-inline-size: 100%;
		inline-size: calc(100vw - 1rem - 2rem);
	}
}

.defer-days {
	justify-content: space-between;
	display: flex;
	margin: .5rem 0;
}

:deep() {
	input.input {
		display: none;
	}

	.flatpickr-calendar {
		margin: 0 auto;
		box-shadow: none;

		@media screen and (max-width: ($defer-task-max-width)) {
			max-inline-size: 100%;
		}

		span {
			inline-size: auto !important;
		}

	}

	.flatpickr-innerContainer {
		@media screen and (max-width: ($defer-task-max-width)) {
			overflow: scroll;
		}
	}
}
</style>
