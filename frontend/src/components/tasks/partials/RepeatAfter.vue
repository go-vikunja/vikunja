<template>
	<div class="control repeat-after-input">
		<div class="button-group mbs-2">
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setRepeatAfter(1, 'days')"
			>
				{{ $t('task.repeat.everyDay') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setRepeatAfter(1, 'weeks')"
			>
				{{ $t('task.repeat.everyWeek') }}
			</XButton>
			<XButton
				variant="secondary"
				class="is-small"
				@click="() => setRepeatAfter(30, 'days')"
			>
				{{ $t('task.repeat.every30d') }}
			</XButton>
		</div>
		<div class="is-flex is-align-items-center mbe-2">
			<label
				for="repeatMode"
				class="is-fullwidth"
			>
				{{ $t('task.repeat.mode') }}:
			</label>
			<div class="control">
				<div class="select">
					<select
						id="repeatMode"
						v-model="task.repeat_mode"
						@change="updateData"
					>
						<option :value="TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT">
							{{ $t('misc.default') }}
						</option>
						<option :value="TASK_REPEAT_MODES.REPEAT_MODE_MONTH">
							{{ $t('task.repeat.monthly') }}
						</option>
						<option :value="TASK_REPEAT_MODES.REPEAT_MODE_FROM_CURRENT_DATE">
							{{ $t('task.repeat.fromCurrentDate') }}
						</option>
					</select>
				</div>
			</div>
		</div>
		<div
			v-if="task.repeat_mode !== TASK_REPEAT_MODES.REPEAT_MODE_MONTH"
			class="is-flex"
		>
			<p class="pis-4">
				{{ $t('task.repeat.each') }}
			</p>
			<div class="field has-addons is-fullwidth">
				<div class="control">
					<input
						v-model="repeatAfter.amount"
						:disabled="disabled || undefined"
						class="input"
						:placeholder="$t('task.repeat.specifyAmount')"
						type="number"
						min="0"
						@change="updateData"
					>
				</div>
				<div class="control">
					<div class="select">
						<select
							v-model="repeatAfter.type"
							:disabled="disabled || undefined"
							@change="updateData"
						>
							<option value="hours">
								{{ $t('task.repeat.hours') }}
							</option>
							<option value="days">
								{{ $t('task.repeat.days') }}
							</option>
							<option value="weeks">
								{{ $t('task.repeat.weeks') }}
							</option>
						</select>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref, reactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import {error} from '@/message'

import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import type {IRepeatAfter} from '@/types/IRepeatAfter'
import type {Task as ITask} from '@/client/generated'
import {createTaskDraft, parseRepeatAfter, repeatAfterToSeconds} from '@/helpers/task'

const props = withDefaults(defineProps<{
	modelValue: ITask | undefined,
	disabled?: boolean
}>(), {
	disabled: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: ITask | undefined],
}>()

const {t} = useI18n({useScope: 'global'})

const task = ref<ITask>(createTaskDraft())
const repeatAfter = reactive({
	amount: 0,
	type: '',
})

watch(
	() => props.modelValue,
	(value: ITask) => {
		task.value = {...value}
		if (typeof value.repeat_after !== 'undefined') {
			Object.assign(repeatAfter, parseRepeatAfter(value.repeat_after))
		}
	},
	{
		immediate: true,
		deep: true,
	},
)

function updateData() {
	if (!task.value || 
		(task.value.repeat_mode === TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT && repeatAfter.amount === 0) ||
		(task.value.repeat_mode === TASK_REPEAT_MODES.REPEAT_MODE_FROM_CURRENT_DATE && repeatAfter.amount === 0)
	) {
		return
	}

	if (task.value.repeat_mode === TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT && repeatAfter.amount < 0) {
		error({message: t('task.repeat.invalidAmount')})
		return
	}

	task.value.repeat_after = repeatAfterToSeconds(repeatAfter as IRepeatAfter)
	emit('update:modelValue', task.value)
}

function setRepeatAfter(amount: number, type: IRepeatAfter['type']) {
	Object.assign(repeatAfter, { amount, type})
	updateData()
}
</script>

<style lang="scss" scoped>
p {
	padding-block-start: 6px;
}

.input {
	min-inline-size: 2rem;
}

.button-group {
	display: flex;
	justify-content: center;

	:deep(.button:not(:last-child)) {
		border-start-end-radius: 0;
		border-end-end-radius: 0;
	}

	:deep(.button:not(:first-child)) {
		border-start-start-radius: 0;
		border-end-start-radius: 0;
	}
}
</style>
