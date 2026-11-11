<template>
	<DatepickerShell
		:open="open"
		:anchor="anchor"
		:sheet-title="$t('input.datepickerRange.date')"
		:selections="selections"
		@update:open="isOpen => $emit('update:open', isOpen)"
	>
		<label class="label">
			{{ $t('input.datepickerRange.date') }}
			<input
				v-model="date"
				class="input"
				type="text"
			>
		</label>

		<CalendarMonth
			:selected="selectedDate"
			@pick="pickDay"
		/>
	</DatepickerShell>
</template>

<script lang="ts" setup>
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import {parseDateOrString} from '@/helpers/time/parseDateOrString'
import {toISOStringOrNull} from '@/helpers/time/toISOStringOrNull'
import {formatDate} from '@/helpers/time/formatDate'

import CalendarMonth from '@/components/input/datepicker/CalendarMonth.vue'
import DatepickerShell from '@/components/date/DatepickerShell.vue'
import {DATE_VALUES} from '@/components/date/dateRanges'

const props = withDefaults(defineProps<{
	modelValue: string | Date | null,
	open?: boolean
	anchor?: HTMLElement | null
}>(), {
	open: false,
	anchor: null,
})

const emit = defineEmits<{
	'update:modelValue': [value: string | Date | null],
	'update:open': [open: boolean],
}>()

const {t} = useI18n({useScope: 'global'})

const date = ref<string|Date|null>('')

watch(
	() => props.modelValue,
	newValue => {
		date.value = newValue
	},
)

const selectedDate = computed<Date | null>(() => {
	const raw = date.value instanceof Date ? toISOStringOrNull(date.value) : date.value
	const parsed = parseDateOrString(raw, null)
	return parsed instanceof Date ? parsed : null
})

function emitChanged() {
	emit('update:modelValue', date.value === '' ? null : date.value)
}

watch(() => date.value, emitChanged)

function pickDay(day: Date) {
	date.value = formatDate(day, 'YYYY-MM-DD HH:mm')
}

function setDate(range: string | null) {
	if (range === null) {
		date.value = ''

		return
	}

	date.value = range
}

const customRangeActive = computed<boolean>(() => {
	return !Object.values(DATE_VALUES).some(d => date.value === d)
})

const selections = computed(() => [
	{
		key: 'custom',
		label: t('misc.custom'),
		active: customRangeActive.value,
		onSelect: () => setDate(null),
	},
	...Object.entries(DATE_VALUES).map(([text, value]) => ({
		key: text,
		label: t(`input.datepickerRange.values.${text}`),
		active: date.value === value,
		onSelect: () => setDate(value),
	})),
])
</script>

<style lang="scss" scoped>
.label {
	margin-block-end: .5rem;
}
</style>
