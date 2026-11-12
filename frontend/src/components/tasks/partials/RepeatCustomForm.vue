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
			class="repeat-custom-preview-region"
			aria-live="polite"
		>
			<template v-if="preview.view.value">
				<div
					v-if="previewList.length > 0"
					class="repeat-custom-preview"
				>
					<span :id="fieldId('preview-label')">{{ t('task.repeat.custom.previewLabel') }}</span>
					<ul :aria-labelledby="fieldId('preview-label')">
						<li
							v-for="row in previewList"
							:key="row.key"
						>
							{{ row.text }}
						</li>
					</ul>
				</div>
				<p
					v-if="preview.state.value === 'stale'"
					class="help repeat-custom-hint repeat-preview-updating"
				>
					{{ t('task.repeat.preview.updating') }}
				</p>
				<p
					v-if="preview.view.value.forecast && previewList.length > 0"
					class="help repeat-custom-hint repeat-preview-forecast"
				>
					{{ t('task.repeat.preview.forecast') }}
				</p>
				<p
					v-if="preview.view.value.reason"
					class="help repeat-custom-hint repeat-preview-reason"
				>
					{{ t(previewReasonKey(preview.view.value.reason)) }}
				</p>
			</template>
			<p
				v-else-if="preview.state.value === 'unavailable'"
				class="help repeat-custom-hint repeat-preview-unavailable"
			>
				{{ t('task.repeat.preview.unavailable') }}
			</p>
			<p
				v-else-if="preview.state.value === 'ruleHint'"
				class="help is-danger repeat-preview-rule-hint"
			>
				{{ t('task.repeat.preview.ruleHint') }}
			</p>
		</div>

		<div class="repeat-custom-actions">
			<XButton
				variant="secondary"
				class="repeat-custom-action"
				:disabled="disabled"
				@click="onBack"
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

<script lang="ts">
import type {TaskRepeat, TaskRepeatWritable} from '@/client/generated'
import {formatInZone} from '@/helpers/rruleDescribe'

// Why the server returned fewer dates than asked; a reason newer than this list gets a generic line.
const PREVIEW_REASON_KEYS: Record<string, string> = {
	count_exhausted: 'task.repeat.reason.count_exhausted',
	until_passed: 'task.repeat.reason.until_passed',
	no_rule: 'task.repeat.reason.no_rule',
	no_due_date: 'task.repeat.reason.no_due_date',
	task_done: 'task.repeat.reason.task_done',
	no_occurrence: 'task.repeat.reason.no_occurrence',
	invalid_rule: 'task.repeat.reason.invalid_rule',
	interval_beyond_horizon: 'task.repeat.reason.interval_beyond_horizon',
	anchor_beyond_horizon: 'task.repeat.reason.anchor_beyond_horizon',
	outside_bounds: 'task.repeat.reason.outside_bounds',
	policy_rejected: 'task.repeat.reason.policy_rejected',
	rule_too_long: 'task.repeat.reason.rule_too_long',
	evaluation_failed: 'task.repeat.reason.evaluation_failed',
	outside_storable_range: 'task.repeat.reason.outside_storable_range',
	no_zone: 'task.repeat.reason.no_zone',
}

const SUB_DAILY = ['hourly', 'minutely', 'secondly']

export type PreviewRow = {key: string, text: string}

export function previewReasonKey(reason: string): string {
	return Object.hasOwn(PREVIEW_REASON_KEYS, reason) ? PREVIEW_REASON_KEYS[reason]! : 'task.repeat.reason.other'
}

// The server reads a rule without the read-only marker, which only ever travels from it.
export function writableRule(repeat: TaskRepeat): TaskRepeatWritable {
	const {read_only: _readOnly, ...rule} = repeat
	return rule
}

function utcOffset(date: Date, zone: string, locale: string): string {
	return new Intl.DateTimeFormat(locale, {timeZone: zone, timeZoneName: 'shortOffset'})
		.formatToParts(date)
		.find(part => part.type === 'timeZoneName')?.value ?? ''
}

function withDuplicatesMarked(texts: string[], mark: (index: number) => string): string[] {
	return texts.map((text, index) => texts.indexOf(text) !== texts.lastIndexOf(text) ? `${text} (${mark(index)})` : text)
}

// Dates are shown in the zone the server answered in. A rule that repeats within a day shows times, and a time the
// clocks pass twice is told apart by its offset.
export function previewRows(dates: string[], zone: string, locale: string, repeat: TaskRepeat | null): PreviewRow[] {
	const instants = dates.map(date => new Date(date))
	const byDay = instants.map(date => formatInZone(date, zone, locale, 'date'))
	const withTime = SUB_DAILY.includes(repeat?.freq ?? '') || new Set(byDay).size < byDay.length
	const texts = withTime
		? withDuplicatesMarked(instants.map(date => formatInZone(date, zone, locale, 'dateTime')), index => utcOffset(instants[index]!, zone, locale))
		: byDay
	return dates.map((date, index) => ({key: date, text: texts[index]!}))
}
</script>

<script setup lang="ts">
import {computed, nextTick, ref, useId} from 'vue'
import {useI18n} from 'vue-i18n'

import FormSelect from '@/components/input/FormSelect.vue'
import CalendarMonth from '@/components/input/datepicker/CalendarMonth.vue'
import XButton from '@/components/input/Button.vue'
import {
	classifyRepeat,
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
import {describeRepeatRule, formatOrdinal, weekdayName, type DescribeOptions} from '@/helpers/rruleDescribe'
import {useOccurrencePreview} from '@/composables/useOccurrencePreview'

const props = withDefaults(defineProps<{
	modelValue: RepeatForm
	dueDate: Date | null
	anchor: Date
	repeatsFromCurrentDate: boolean
	zone: string
	weekStart: number
	taskId: number
	/** Fixed when the editor opened; the preview of a task without a due date asks the server for this time. */
	initialDueTime: string | null
	done?: boolean
	/** The repeat as stored, so a count-only raise the server would ignore is refused here. */
	storedRepeat?: TaskRepeat | null
	disabled?: boolean
}>(), {
	storedRepeat: null,
	done: false,
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

function onBack() {
	if (!props.disabled) {
		emit('back')
	}
}

// The button is disabled too, but a click can still reach this through the component's own event.
function onDone() {
	if (props.disabled || errors.value.length > 0) {
		return
	}
	emit('done')
}

// Each form asks for its own preview, so the draft's answers never replace the summary's.
const preview = useOccurrencePreview(() => errors.value.length > 0 ? null : {
	repeat: writableRule(repeat.value),
	repeatClass: classifyRepeat(repeat.value, props.dueDate),
	dueDate: props.dueDate,
	initialDueTime: props.initialDueTime,
	taskId: props.taskId,
	done: props.done,
	repeatsFromCurrentDate: props.repeatsFromCurrentDate,
})

const previewList = computed(() => {
	const view = preview.view.value
	return view === null ? [] : previewRows(view.dates, view.zone, locale.value, repeat.value)
})
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

.repeat-custom-preview-region {
	display: flex;
	flex-direction: column;
	gap: .25rem;

	&:empty {
		display: none;
	}
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
