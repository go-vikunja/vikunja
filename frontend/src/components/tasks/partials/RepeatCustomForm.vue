<template>
	<div class="repeat-custom-form">
		<div class="repeat-custom-every">
			<label
				:for="fieldId('interval')"
				class="repeat-custom-every-label"
			>
				{{ t('task.repeat.custom.every') }}
			</label>
			<input
				:id="fieldId('interval')"
				class="input repeat-custom-number"
				type="number"
				min="1"
				step="1"
				inputmode="numeric"
				:value="numberValue(modelValue.interval)"
				:disabled="disabled || undefined"
				:aria-invalid="errorFor('interval') ? true : undefined"
				:aria-describedby="describedBy(errorFor('interval') ? errorId('interval') : null)"
				@input="update({interval: parseNumber($event)})"
			>
			<FormSelect
				class="repeat-custom-unit"
				:model-value="modelValue.unit"
				:options="unitOptions"
				:disabled="disabled"
				:aria-label="t('task.repeat.custom.unit')"
				@update:modelValue="value => update({unit: value as RepeatUnit})"
			/>
			<p
				v-if="errorFor('interval')"
				:id="errorId('interval')"
				class="help is-danger repeat-custom-error"
			>
				{{ errorText('interval') }}
			</p>
		</div>

		<fieldset
			v-if="modelValue.unit === 'week'"
			class="repeat-custom-group"
			:aria-describedby="describedBy(weekdaysEmpty ? fieldId('weekday-hint') : null)"
		>
			<legend class="repeat-custom-legend">
				{{ t('task.repeat.custom.weekdays') }}
			</legend>
			<div class="repeat-custom-weekdays">
				<button
					v-for="weekday in orderedWeekdays"
					:key="weekday.code"
					type="button"
					class="repeat-custom-weekday"
					:class="{'is-pressed': weekday.pressed}"
					:aria-pressed="weekday.pressed"
					:aria-label="weekday.longName"
					:disabled="disabled || undefined"
					@click="toggleWeekday(weekday.code)"
				>
					{{ weekday.shortName }}
				</button>
			</div>
			<p
				v-if="weekdaysEmpty"
				:id="fieldId('weekday-hint')"
				class="help repeat-custom-hint repeat-custom-weekday-hint"
			>
				{{ t('task.repeat.custom.weekdayFallback', {weekday: weekdayName(anchorDay.weekday, locale)}) }}
			</p>
		</fieldset>

		<fieldset
			v-if="modelValue.unit === 'month'"
			class="repeat-custom-group"
		>
			<legend class="repeat-custom-legend">
				{{ t('task.repeat.custom.monthMode') }}
			</legend>
			<div class="repeat-custom-option">
				<label class="radio repeat-custom-choice">
					<input
						type="radio"
						:name="fieldId('month-mode')"
						value="day"
						:checked="modelValue.monthMode === 'day'"
						:disabled="disabled || undefined"
						@change="update({monthMode: 'day'})"
					>
					{{ t('task.repeat.custom.onDay') }}
				</label>
				<FormSelect
					class="repeat-custom-month-day"
					:model-value="modelValue.monthDay ?? anchorDay.day"
					:options="monthDayOptions"
					:disabled="disabled"
					:error="errorText('monthDay')"
					:aria-label="t('task.repeat.custom.monthDay')"
					@update:modelValue="value => update({monthMode: 'day', monthDay: Number(value)})"
				/>
			</div>
			<div class="repeat-custom-option">
				<label class="radio repeat-custom-choice">
					<input
						type="radio"
						:name="fieldId('month-mode')"
						value="nthWeekday"
						:checked="modelValue.monthMode === 'nthWeekday'"
						:disabled="disabled || undefined"
						@change="update({monthMode: 'nthWeekday'})"
					>
					{{ t('task.repeat.custom.onThe') }}
				</label>
				<FormSelect
					class="repeat-custom-ordinal"
					:model-value="modelValue.ordinal"
					:options="ordinalOptions"
					:disabled="disabled"
					:aria-label="t('task.repeat.custom.ordinal')"
					@update:modelValue="value => update({monthMode: 'nthWeekday', ordinal: Number(value) as RepeatOrdinal})"
				/>
				<FormSelect
					class="repeat-custom-month-weekday"
					:model-value="modelValue.monthWeekday"
					:options="monthWeekdayOptions"
					:disabled="disabled"
					:aria-label="t('task.repeat.custom.monthWeekday')"
					@update:modelValue="value => update({monthMode: 'nthWeekday', monthWeekday: value as WeekdayCode})"
				/>
			</div>
			<div class="repeat-custom-option">
				<label class="radio repeat-custom-choice">
					<input
						type="radio"
						:name="fieldId('month-mode')"
						value="lastDay"
						:checked="modelValue.monthMode === 'lastDay'"
						:disabled="disabled || undefined"
						@change="update({monthMode: 'lastDay'})"
					>
					{{ t('task.repeat.custom.onLastDay') }}
				</label>
			</div>
		</fieldset>

		<fieldset class="repeat-custom-group">
			<legend class="repeat-custom-legend">
				{{ t('task.repeat.custom.end') }}
			</legend>
			<div class="repeat-custom-option">
				<label class="radio repeat-custom-choice">
					<input
						type="radio"
						:name="fieldId('end')"
						value="never"
						:checked="modelValue.end === 'never'"
						:disabled="disabled || undefined"
						@change="update({end: 'never'})"
					>
					{{ t('task.repeat.custom.endNever') }}
				</label>
			</div>
			<div class="repeat-custom-option">
				<label class="radio repeat-custom-choice">
					<input
						type="radio"
						:name="fieldId('end')"
						value="count"
						:checked="modelValue.end === 'count'"
						:disabled="disabled || undefined"
						@change="update({end: 'count'})"
					>
					{{ t('task.repeat.custom.endAfter') }}
				</label>
				<input
					:id="fieldId('count')"
					class="input repeat-custom-number"
					type="number"
					min="1"
					step="1"
					inputmode="numeric"
					:value="numberValue(modelValue.count)"
					:disabled="disabled || undefined"
					:aria-label="t('task.repeat.custom.count')"
					:aria-invalid="errorFor('count') ? true : undefined"
					:aria-describedby="describedBy(fieldId('count-hint'), errorFor('count') ? errorId('count') : null)"
					@input="update({end: 'count', count: parseNumber($event)})"
				>
				<span class="repeat-custom-times">{{ t('task.repeat.custom.times', pluralCount(modelValue.count)) }}</span>
				<p
					:id="fieldId('count-hint')"
					class="help repeat-custom-hint"
				>
					{{ t('task.repeat.custom.countIncludesCurrent') }}
				</p>
				<p
					v-if="errorFor('count')"
					:id="errorId('count')"
					class="help is-danger repeat-custom-error"
				>
					{{ errorText('count') }}
				</p>
			</div>
			<div class="repeat-custom-option">
				<label class="radio repeat-custom-choice">
					<input
						ref="untilRadio"
						type="radio"
						:name="fieldId('end')"
						value="until"
						:checked="modelValue.end === 'until'"
						:disabled="disabled || undefined"
						:aria-invalid="errorFor('untilDate') ? true : undefined"
						:aria-describedby="describedBy(errorFor('untilDate') ? errorId('untilDate') : null)"
						@change="update({end: 'until'})"
					>
					{{ t('task.repeat.custom.endOn') }}
				</label>
				<template v-if="modelValue.end === 'until' && untilText">
					<span class="repeat-custom-until-text">{{ untilText }}</span>
					<XButton
						variant="tertiary"
						class="repeat-custom-clear"
						:disabled="disabled"
						@click="clearUntil"
					>
						{{ t('task.repeat.custom.clearEndDate') }}
					</XButton>
				</template>
				<p
					v-if="errorFor('untilDate')"
					:id="errorId('untilDate')"
					class="help is-danger repeat-custom-error"
				>
					{{ errorText('untilDate') }}
				</p>
			</div>
			<!-- The calendar has no disabled state of its own, so inert keeps it out of reach while the form is disabled. -->
			<div
				v-if="modelValue.end === 'until'"
				class="repeat-custom-until-picker"
				role="group"
				:aria-label="t('task.repeat.custom.untilDate')"
				:aria-describedby="describedBy(errorFor('untilDate') ? errorId('untilDate') : null)"
				:inert="disabled || undefined"
			>
				<CalendarMonth
					large
					:selected="untilPickerDate"
					@pick="onPickUntil"
				/>
			</div>
		</fieldset>

		<p
			class="repeat-custom-summary"
			aria-live="polite"
		>
			{{ summary }}
		</p>
		<div
			v-if="previewText.length > 0"
			class="repeat-custom-preview"
		>
			<span :id="fieldId('preview-label')">{{ t('task.repeat.custom.previewLabel') }}</span>
			<ul :aria-labelledby="fieldId('preview-label')">
				<li
					v-for="date in previewText"
					:key="date"
				>
					{{ date }}
				</li>
			</ul>
		</div>
		<p
			v-else-if="preview.hiddenReason === 'noZone'"
			class="help repeat-custom-hint repeat-custom-preview-hint"
		>
			{{ t('task.repeat.custom.noPreview') }}
		</p>

		<div class="repeat-custom-actions">
			<XButton
				variant="secondary"
				class="repeat-custom-action"
				@click="emit('back')"
			>
				{{ t('task.repeat.custom.back') }}
			</XButton>
			<XButton
				class="repeat-custom-action"
				:disabled="disabled || errors.length > 0"
				@click="onDone"
			>
				{{ t('task.repeat.custom.done') }}
			</XButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, ref, useId} from 'vue'
import {useI18n} from 'vue-i18n'

import type {TaskRepeat} from '@/client/generated'
import FormSelect from '@/components/input/FormSelect.vue'
import CalendarMonth from '@/components/input/datepicker/CalendarMonth.vue'
import XButton from '@/components/input/Button.vue'
import {
	dayInZone,
	formToRepeat,
	normalizeRepeatForm,
	pickerDateToUntilDate,
	untilDateToPickerDate,
	validateRepeatForm,
	type RepeatForm,
	type RepeatOrdinal,
	type RepeatUnit,
	type WeekdayCode,
} from '@/helpers/rrule'
import {previewDates} from '@/helpers/rrulePreview'
import {describeRepeatRule, formatInZone, formatOrdinal, weekdayName, type DescribeOptions} from '@/helpers/rruleDescribe'

const props = withDefaults(defineProps<{
	modelValue: RepeatForm
	dueDate: Date | null
	anchor: Date
	repeatsFromCurrentDate: boolean
	zone: string
	previewZone: string | null
	weekStart: number
	/** The repeat as stored, so a count-only raise the server would ignore is refused here. */
	storedRepeat?: TaskRepeat | null
	disabled?: boolean
}>(), {
	storedRepeat: null,
	disabled: false,
})
const emit = defineEmits<{
	'update:modelValue': [value: RepeatForm]
	done: []
	back: []
}>()
const {t, locale} = useI18n({useScope: 'global'})

const translate: DescribeOptions['t'] = (key, params, plural) => plural === undefined
	? t(key, params ?? {})
	: t(key, params ?? {}, plural)

const uid = useId()
const untilRadio = ref<HTMLInputElement | null>(null)

const UNITS: readonly RepeatUnit[] = ['hour', 'day', 'week', 'month', 'year']
const WEEKDAYS_FROM_MONDAY: readonly WeekdayCode[] = ['mo', 'tu', 'we', 'th', 'fr', 'sa', 'su']
const WEEKDAYS_FROM_SUNDAY: readonly WeekdayCode[] = ['su', 'mo', 'tu', 'we', 'th', 'fr', 'sa']
const ORDINALS: readonly RepeatOrdinal[] = [1, 2, 3, 4, -1]

function fieldId(name: string): string {
	return `${uid}-${name}`
}

function errorId(field: keyof RepeatForm): string {
	return fieldId(`${field}-error`)
}

function describedBy(...ids: (string | null)[]): string | undefined {
	const present = ids.filter(id => id !== null)
	return present.length > 0 ? present.join(' ') : undefined
}

function update(changes: Partial<RepeatForm>) {
	emit('update:modelValue', {...props.modelValue, ...changes})
}

// An empty field stays NaN in the draft so validation reports it instead of silently becoming 0.
function parseNumber(event: Event): number {
	const value = (event.target as HTMLInputElement).value
	return value === '' ? Number.NaN : Number(value)
}

function numberValue(value: number): number | string {
	return Number.isFinite(value) ? value : ''
}

function pluralCount(value: number): number {
	return Number.isFinite(value) ? value : 1
}

const anchorDay = computed(() => dayInZone(props.anchor, props.zone))

const unitOptions = computed(() => UNITS.map(unit => ({
	value: unit,
	label: t(`task.repeat.unit.${unit}`, pluralCount(props.modelValue.interval)),
})))

const weekdaysFromWeekStart = computed(() => {
	const start = ((props.weekStart % 7) + 7) % 7
	return [...WEEKDAYS_FROM_SUNDAY.slice(start), ...WEEKDAYS_FROM_SUNDAY.slice(0, start)]
})

const orderedWeekdays = computed(() => {
	const selected = props.modelValue.weekdays ?? [anchorDay.value.weekday]
	return weekdaysFromWeekStart.value.map(code => ({
		code,
		pressed: selected.includes(code),
		shortName: weekdayName(code, locale.value, 'short'),
		longName: weekdayName(code, locale.value, 'long'),
	}))
})

const weekdaysEmpty = computed(() => props.modelValue.weekdays !== null && props.modelValue.weekdays.length === 0)

function toggleWeekday(code: WeekdayCode) {
	const current = props.modelValue.weekdays ?? [anchorDay.value.weekday]
	const next = current.includes(code) ? current.filter(day => day !== code) : [...current, code]
	update({weekdays: WEEKDAYS_FROM_MONDAY.filter(day => next.includes(day))})
}

const monthDayOptions = Array.from({length: 31}, (_, i) => i + 1)

const ordinalOptions = computed(() => ORDINALS.map(ordinal => ({
	value: ordinal,
	label: ordinal === -1 ? t('task.repeat.custom.lastOrdinal') : formatOrdinal(ordinal, {t: translate, locale: locale.value}),
})))

const monthWeekdayOptions = computed(() => weekdaysFromWeekStart.value.map(code => ({
	value: code,
	label: weekdayName(code, locale.value, 'long'),
})))

function ymdInZone(date: Date): string {
	const {year, month, day} = dayInZone(date, props.zone)
	return `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
}

const untilText = computed(() => {
	const {untilDate, untilInstant} = props.modelValue
	if (!untilDate) {
		return ''
	}
	if (untilInstant) {
		const stored = new Date(untilInstant)
		// An end date picked here is the last second of its day; any other stored time is shown, as it is kept verbatim.
		const isEndOfDay = ymdInZone(new Date(stored.getTime() + 1000)) !== ymdInZone(stored)
		if (!isEndOfDay && ymdInZone(stored) === untilDate) {
			return formatInZone(stored, props.zone, locale.value, 'dateTimeYear')
		}
	}
	// untilDate is already a calendar date in the recurrence zone; noon UTC names that same date.
	const [year, month, day] = untilDate.split('-').map(Number) as [number, number, number]
	return formatInZone(new Date(Date.UTC(year, month - 1, day, 12)), 'UTC', locale.value, 'monthDayYear')
})

// Keyed on the date string: a fresh Date on every draft change would make the calendar jump back to it.
const untilDate = computed(() => props.modelValue.untilDate)
const untilPickerDate = computed(() => untilDateToPickerDate(untilDate.value))

function onPickUntil(date: Date) {
	if (props.disabled) {
		return
	}
	update({end: 'until', untilDate: pickerDateToUntilDate(date)})
}

// The clear button disappears with the date, so focus moves to the radio, which announces the required-date error.
async function clearUntil() {
	if (props.disabled) {
		return
	}
	update({untilDate: null})
	await nextTick()
	untilRadio.value?.focus()
}

const errors = computed(() => validateRepeatForm(props.modelValue, props.dueDate, props.zone, props.storedRepeat))

function errorFor(field: keyof RepeatForm) {
	return errors.value.find(error => error.field === field) ?? null
}

function errorText(field: keyof RepeatForm): string | null {
	const error = errorFor(field)
	return error === null ? null : t(error.key, error.params ?? {})
}

const repeat = computed(() => formToRepeat(normalizeRepeatForm(props.modelValue), props.dueDate, props.zone))

const summary = computed(() => describeRepeatRule(repeat.value, {
	t: translate,
	locale: locale.value,
	weekStart: props.weekStart,
	zone: props.zone,
	dueDate: props.anchor,
}))

// The button is disabled too, but a click can still reach this through the component's own event.
function onDone() {
	if (props.disabled || errors.value.length > 0) {
		return
	}
	emit('done')
}

const preview = computed(() => previewDates(repeat.value, props.dueDate, new Date(), props.repeatsFromCurrentDate, props.previewZone))

const previewText = computed(() => (preview.value.dates ?? []).map(date => formatInZone(date, props.zone, locale.value, 'date')))
</script>

<style lang="scss" scoped>
$hit-size: 2.5rem;

.repeat-custom-form {
	container-type: inline-size;
	display: flex;
	flex-direction: column;
	gap: .75rem;
	max-inline-size: 100%;
	// Long compound words in some languages would otherwise push the sheet sideways.
	overflow-wrap: anywhere;
}

.repeat-custom-every {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .5rem;
}

.repeat-custom-every-label {
	display: flex;
	align-items: center;
	min-block-size: $hit-size;
}

.repeat-custom-number {
	flex: 0 0 4.5rem;
	min-block-size: $hit-size;
	inline-size: 4.5rem;
	min-inline-size: 0;
}

// FormSelect renders a fragment and forwards attributes to its <select>, so its parts are reached with :deep.
.repeat-custom-form :deep(.select) {
	min-inline-size: 0;
	max-inline-size: 100%;
	min-block-size: $hit-size;
}

.repeat-custom-form :deep(select) {
	min-block-size: $hit-size;
	max-inline-size: 100%;
}

// Long unit names move the select onto its own line instead of overflowing the row.
.repeat-custom-every :deep(.select) {
	flex: 1 1 7rem;
}

.repeat-custom-error,
.repeat-custom-hint,
.repeat-custom-option :deep(.help) {
	flex-basis: 100%;
	margin: 0;
}

// Bulma's is-danger text is --danger, 3.46:1 on white; FormSelect's own errors are reached with :deep.
.repeat-custom-form :deep(.help.is-danger) {
	color: var(--danger-text);
}

.repeat-custom-hint {
	color: var(--grey-600);
}

.repeat-custom-group {
	display: flex;
	flex-direction: column;
	gap: .25rem;
	min-inline-size: 0;
	margin: 0;
	padding: 0;
	border: 0;
}

.repeat-custom-legend {
	margin-block-end: .25rem;
	padding: 0;
	font-weight: 600;
}

.repeat-custom-weekdays {
	display: grid;
	grid-template-columns: repeat(4, minmax($hit-size, 1fr));
	gap: .25rem;
}

@container (min-width: 21rem) {
	.repeat-custom-weekdays {
		grid-template-columns: repeat(7, minmax($hit-size, 1fr));
	}
}

.repeat-custom-weekday {
	min-block-size: $hit-size;
	min-inline-size: $hit-size;
	padding: 0 .25rem;
	border: 1px solid var(--grey-300);
	border-radius: $radius;
	background: var(--white);
	color: var(--text);
	font: inherit;
	cursor: pointer;

	&.is-pressed {
		border-color: var(--primary);
		background: var(--primary);
		color: var(--white);
	}

	&:focus-visible {
		outline: 2px solid var(--primary);
		outline-offset: 2px;
	}

	&:disabled {
		cursor: not-allowed;
		opacity: .6;
	}
}

.repeat-custom-option {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .25rem .5rem;
	min-inline-size: 0;
}

.repeat-custom-choice {
	display: inline-flex;
	align-items: center;
	gap: .5rem;
	min-block-size: $hit-size;
	min-inline-size: $hit-size;
	margin: 0;
}

.repeat-custom-until-text {
	color: var(--grey-600);
}

.repeat-custom-clear {
	min-block-size: $hit-size;
	min-inline-size: $hit-size;
}

.repeat-custom-until-picker {
	min-inline-size: 0;
}

// The shared calendar's month controls are sized for a pointer; the form needs touch-sized ones.
.repeat-custom-until-picker :deep(.calendar-month__nav-button) {
	inline-size: $hit-size;
	block-size: $hit-size;
}

.repeat-custom-until-picker :deep(.calendar-month__month-select),
.repeat-custom-until-picker :deep(.calendar-month__year-input) {
	min-block-size: $hit-size;
	min-inline-size: $hit-size;
}

.repeat-custom-until-picker :deep(.calendar-month__day) {
	min-block-size: $hit-size;
	min-inline-size: $hit-size;
}

.repeat-custom-summary {
	margin: 0;
	font-weight: 600;
}

.repeat-custom-preview {
	display: flex;
	flex-wrap: wrap;
	gap: .25rem .5rem;
	color: var(--grey-600);

	ul {
		display: flex;
		flex-wrap: wrap;
		gap: .25rem .75rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}
}

.repeat-custom-actions {
	display: flex;
	justify-content: flex-end;
	gap: .5rem;
}

.repeat-custom-action {
	min-block-size: $hit-size;
	min-inline-size: $hit-size;
}
</style>
