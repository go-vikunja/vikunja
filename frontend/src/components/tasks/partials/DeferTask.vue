<template>
	<div
		:class="{ 'is-loading': update.isPending.value }"
		class="defer-task loading-container"
		@click.stop
		@mousedown.stop
		@pointerdown.stop
		@touchstart.stop
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
		<DatepickerInline
			v-model="dueDate"
			:show-shortcuts="false"
			@update:modelValue="onPickerUpdate"
		/>
	</div>
</template>

<script setup lang="ts">
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {ref, watch, onBeforeUnmount} from 'vue'
import {useDebounceFn} from '@vueuse/core'

import DatepickerInline from '@/components/input/DatepickerInline.vue'

import {useUpdateTaskMutation} from '@/client/queries/taskMutations'
import type {Task as ITask} from '@/client/generated'

const props = defineProps<{
	modelValue: ITask,
}>()

const emit = defineEmits<{
	'update:modelValue': [value: ITask]
}>()

const update = useUpdateTaskMutation()
const task = ref<ITask>()

// We're saving the due date separately to prevent null errors in very short periods where the task is null.
const dueDate = ref<Date | null>(null)
const lastValue = ref<Date | null>()

watch(
	() => props.modelValue,
	(value) => {
		task.value = { ...value }
		dueDate.value = parseDateOrNull(value.due_date)
		lastValue.value = dueDate.value
	},
	{immediate: true},
)

function deferDays(days: number) {
	debouncedUpdateDueDate.cancel()
	const deferred = new Date(dueDate.value ?? new Date())
	deferred.setDate(deferred.getDate() + days)
	dueDate.value = deferred
	updateDueDate()
}

const debouncedUpdateDueDate = useDebounceFn(updateDueDate, 500)

function onPickerUpdate() {
	debouncedUpdateDueDate()
}

onBeforeUnmount(() => debouncedUpdateDueDate.flush())

async function updateDueDate() {
	if (!dueDate.value) {
		return
	}

	if (+new Date(dueDate.value) === +lastValue.value) {
		return
	}

	const newTask = await update.mutateAsync({
		...task.value,
		id: task.value!.id!,
		due_date: new Date(dueDate.value).toISOString(),
	})
	lastValue.value = parseDateOrNull(newTask.due_date)
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

	.bottom-sheet & {
		max-inline-size: none;
		padding: 0 1rem 1rem;

		> .label {
			display: none;
		}
	}
}

.defer-days {
	justify-content: space-between;
	display: flex;
	gap: .5rem;
	margin: .5rem 0;

	.bottom-sheet & {
		margin-block-start: 0;

		> * {
			flex: 1;
		}
	}
}
</style>
