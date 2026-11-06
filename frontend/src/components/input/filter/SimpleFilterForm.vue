<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {isDateMath, type SimpleFilter} from '@/helpers/simpleFilter'
import type {Priority} from '@/constants/priorities'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import DatepickerWithRange from '@/components/date/DatepickerWithRange.vue'
import PrioritySelect from '@/components/tasks/partials/PrioritySelect.vue'
import FilterLabelSelect from './FilterLabelSelect.vue'
import FilterAssigneeSelect from './FilterAssigneeSelect.vue'
import FilterProjectSelect from './FilterProjectSelect.vue'
import XButton from '@/components/input/Button.vue'

const props = defineProps<{
	projectId?: number
}>()
const model = defineModel<SimpleFilter>({required: true})
const {t} = useI18n({useScope: 'global'})

function update(patch: Partial<SimpleFilter>) {
	model.value = {
		...model.value,
		...patch,
	}
}

const dates = ref<{
	dateFrom: Date | string | null
	dateTo: Date | string | null
}>({
	dateFrom: null,
	dateTo: null,
})
watch(() => model.value.due.range, range => {
	if (serializeDate(dates.value.dateFrom) === (range?.from ?? '') &&
		serializeDate(dates.value.dateTo) === (range?.to ?? '')) {
		return
	}
	dates.value = {
		dateFrom: range?.from || null,
		dateTo: range?.to || null,
	}
}, {immediate: true})

function serializeDate(value: Date | string | null): string {
	if (!value) {
		return ''
	}
	if (typeof value === 'string' && value.startsWith('now')) {
		return isDateMath(value) ? value : ''
	}
	const date = new Date(value)
	return Number.isNaN(date.getTime()) ? '' : date.toISOString()
}

function setDates(value: typeof dates.value) {
	dates.value = value
	const from = serializeDate(value.dateFrom)
	const to = serializeDate(value.dateTo)
	if (!value.dateFrom && !value.dateTo) {
		update({due: {
			...model.value.due,
			range: null,
		}})
		return
	}
	if ((!from || !to) && !(model.value.due.includeOverdue && to)) {
		return
	}
	update({due: {
		...model.value.due,
		range: {
			from,
			to,
		},
	}})
}

function setOverdue(includeOverdue: boolean) {
	const from = serializeDate(dates.value.dateFrom)
	const to = serializeDate(dates.value.dateTo)
	update({due: {
		includeOverdue,
		range: to && (from || includeOverdue) ? {
			from,
			to,
		} : null,
	}})
}

const priority = computed({
	get: () => model.value.minPriority ?? 0,
	set: (value: number) => update({minPriority: value === 0 ? null : value as Priority}),
})
</script>

<template>
	<div class="simple-filter-form">
		<div class="field">
			<FancyCheckbox
				:model-value="model.showDone"
				@update:modelValue="update({showDone: $event})"
			>
				{{ t('filters.simple.showDone') }}
			</FancyCheckbox>
		</div>
		<div class="field">
			<div class="label">
				{{ t('task.attributes.dueDate') }}
			</div>
			<DatepickerWithRange
				:model-value="dates"
				@update:modelValue="setDates"
			>
				<template #trigger="{toggle, buttonText}">
					<XButton
						variant="secondary"
						@click.prevent="toggle()"
					>
						{{ buttonText }}
					</XButton>
				</template>
			</DatepickerWithRange>
			<XButton
				v-if="dates.dateFrom || dates.dateTo"
				variant="tertiary"
				@click="dates = {dateFrom: null, dateTo: null}; update({due: {...model.due, range: null}})"
			>
				{{ t('input.datepicker.clear') }}
			</XButton>
		</div>
		<div class="field">
			<FancyCheckbox
				:model-value="model.due.includeOverdue"
				@update:modelValue="setOverdue"
			>
				{{ t('task.show.overdue') }}
			</FancyCheckbox>
		</div>
		<div class="field">
			<div class="label">
				{{ t('task.attributes.labels') }}
			</div>
			<FilterLabelSelect
				:model-value="model.labelIds"
				@update:modelValue="update({labelIds: $event})"
			/>
		</div>
		<div class="field">
			<div class="label">
				{{ t('task.attributes.assignees') }}
			</div>
			<FilterAssigneeSelect
				:model-value="model.assignees"
				:project-id="props.projectId"
				@update:modelValue="update({assignees: $event})"
			/>
		</div>
		<div
			v-if="!props.projectId"
			class="field"
		>
			<div class="label">
				{{ t('project.projects') }}
			</div>
			<FilterProjectSelect
				:model-value="model.projectIds"
				@update:modelValue="update({projectIds: $event})"
			/>
		</div>
		<div class="field">
			<div class="label">
				{{ t('filters.simple.minPriority') }}
			</div>
			<PrioritySelect v-model="priority" />
		</div>
	</div>
</template>
