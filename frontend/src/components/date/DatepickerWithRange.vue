<template>
	<DatepickerShell
		v-model:open="isPopupOpen"
		:sheet-title="$t('misc.dateRange')"
		:selections="selections"
	>
		<template #trigger="{toggle}">
			<slot
				name="trigger"
				:toggle="toggle"
				:button-text="buttonText"
			/>
		</template>
		<template #default>
			<div class="date-fields">
				<label class="label">
					{{ $t('input.datepickerRange.from') }}
					<input
						v-model="from"
						class="input"
						type="text"
					>
				</label>
				<label class="label">
					{{ $t('input.datepickerRange.to') }}
					<input
						v-model="to"
						class="input"
						type="text"
					>
				</label>
			</div>

			<CalendarMonth
				mode="range"
				:range-start="rangeStart"
				:range-end="rangeEnd"
				@pick="pickDay"
			/>
		</template>
	</DatepickerShell>
</template>

<script lang="ts" setup>
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import {parseDateOrString} from '@/helpers/time/parseDateOrString'
import {formatDate} from '@/helpers/time/formatDate'

import {useRangePick} from '@/composables/useRangePick'

import CalendarMonth from '@/components/input/datepicker/CalendarMonth.vue'
import DatepickerShell from '@/components/date/DatepickerShell.vue'
import {DATE_RANGES} from '@/components/date/dateRanges'

const props = withDefaults(defineProps<{
	// null for a side that's been cleared (the Custom option) — emitted, so accepted too.
	modelValue?: {
		dateFrom: Date | string | null,
		dateTo: Date | string | null,
	},
}>(), {
	modelValue: () => ({dateFrom: null, dateTo: null}),
})

const emit = defineEmits<{
	'update:modelValue': [value: {
		dateFrom: Date | string | null,
		dateTo: Date | string | null
	}]
}>()

const {t} = useI18n({useScope: 'global'})

const from = ref('')
const to = ref('')

watch(
	() => props.modelValue,
	newValue => {
		from.value = typeof newValue.dateFrom === 'string' ? newValue.dateFrom : (newValue.dateFrom?.toISOString() ?? '')
		to.value = typeof newValue.dateTo === 'string' ? newValue.dateTo : (newValue.dateTo?.toISOString() ?? '')
	},
	{immediate: true},
)

function emitChanged() {
	const args = {
		dateFrom: from.value === '' ? null : from.value,
		dateTo: to.value === '' ? null : to.value,
	}
	emit('update:modelValue', args)
}

watch(() => from.value, emitChanged)
watch(() => to.value, emitChanged)

function asDate(value: string): Date | null {
	const parsed = parseDateOrString(value, null)
	return parsed instanceof Date ? parsed : null
}

const isPopupOpen = ref(false)

const {rangeStart, rangeEnd, pickDay, reset: resetPendingRange} = useRangePick(
	() => asDate(from.value),
	() => asDate(to.value),
	(start, end) => {
		// Inclusive end: a range ending "Sep 22" should cover the whole 22nd.
		from.value = formatDate(start, 'YYYY-MM-DD 00:00')
		to.value = formatDate(end, 'YYYY-MM-DD 23:59')
	},
	isPopupOpen,
)

function setDateRange(range: string[] | null) {
	resetPendingRange()
	if (range === null) {
		from.value = ''
		to.value = ''

		return
	}

	from.value = range[0] ?? ''
	to.value = range[1] ?? ''
}

const customRangeActive = computed<boolean>(() => {
	return !Object.values(DATE_RANGES).some(range => from.value === range[0] && to.value === range[1])
})

const selections = computed(() => [
	{
		key: 'custom',
		label: t('misc.custom'),
		active: customRangeActive.value,
		onSelect: () => setDateRange(null),
	},
	...Object.entries(DATE_RANGES).map(([text, range]) => ({
		key: text,
		label: t(`input.datepickerRange.ranges.${text}`),
		active: from.value === range[0] && to.value === range[1],
		onSelect: () => setDateRange([...range]),
	})),
])

const buttonText = computed<string>(() => {
	if (from.value === '' || to.value === '') {
		return t('task.show.select')
	}

	// Show the preset's name when the range matches one, rather than the raw datemath.
	const preset = Object.entries(DATE_RANGES).find(
		([, range]) => from.value === range[0] && to.value === range[1],
	)
	if (preset) {
		return t(`input.datepickerRange.ranges.${preset[0]}`)
	}

	return t('input.datepickerRange.fromto', {
		from: from.value,
		to: to.value,
	})
})
</script>

<style lang="scss" scoped>
.date-fields {
	display: flex;
	gap: .5rem;
	margin-block-end: .5rem;

	.label {
		flex: 1;
	}
}
</style>
