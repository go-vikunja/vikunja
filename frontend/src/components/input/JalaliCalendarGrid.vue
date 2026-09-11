<template>
	<div
		class="jalali-calendar"
		dir="rtl"
		role="group"
		:aria-label="monthTitle"
	>
		<div class="jalali-calendar__nav">
			<button
				type="button"
				class="jalali-calendar__nav-button"
				:aria-label="t('input.datepicker.prevYear')"
				:disabled="disabled"
				@click="stepYear(-1)"
			>
				»
			</button>
			<button
				type="button"
				class="jalali-calendar__nav-button"
				:aria-label="t('input.datepicker.prevMonth')"
				:disabled="disabled"
				@click="stepMonth(-1)"
			>
				›
			</button>
			<div class="jalali-calendar__title">
				{{ monthTitle }}
			</div>
			<button
				type="button"
				class="jalali-calendar__nav-button"
				:aria-label="t('input.datepicker.nextMonth')"
				:disabled="disabled"
				@click="stepMonth(1)"
			>
				‹
			</button>
			<button
				type="button"
				class="jalali-calendar__nav-button"
				:aria-label="t('input.datepicker.nextYear')"
				:disabled="disabled"
				@click="stepYear(1)"
			>
				«
			</button>
		</div>

		<div class="jalali-calendar__weekdays">
			<span
				v-for="(name, index) in weekdayNames"
				:key="index"
				class="jalali-calendar__weekday"
			>
				{{ name }}
			</span>
		</div>

		<div class="jalali-calendar__days">
			<span
				v-for="blank in leadingBlanks"
				:key="`blank-${blank}`"
				class="jalali-calendar__blank"
			/>
			<button
				v-for="cell in cells"
				:key="cell.day"
				ref="dayButtons"
				type="button"
				class="jalali-calendar__day"
				:class="{
					'is-selected': cell.selected,
					'is-today': cell.today,
					'is-range-start': cell.rangeStart,
					'is-range-end': cell.rangeEnd,
					'is-in-range': cell.inRange,
					'is-hovered': cell.hovered,
				}"
				:disabled="disabled || cell.disabled"
				:aria-label="cell.ariaLabel"
				:aria-pressed="mode === 'single' ? cell.selected : undefined"
				:aria-current="cell.today ? 'date' : undefined"
				@click="selectDay(cell.day)"
				@mouseenter="hoverDay(cell.day)"
				@keydown="onDayKeydown($event, cell.day)"
			>
				{{ cell.label }}
			</button>
		</div>

		<div
			v-if="enableTime"
			class="jalali-calendar__time"
		>
			<input
				v-model.number="hoursModel"
				class="jalali-calendar__time-input"
				type="number"
				:min="time24hr ? 0 : 1"
				:max="time24hr ? 23 : 12"
				:disabled="disabled"
				:aria-label="t('input.datepicker.hour')"
				@change="emitWithTime"
			>
			<span class="jalali-calendar__time-separator">:</span>
			<input
				v-model.number="minutesModel"
				class="jalali-calendar__time-input"
				type="number"
				min="0"
				max="59"
				:disabled="disabled"
				:aria-label="t('input.datepicker.minute')"
				@change="emitWithTime"
			>
			<button
				v-if="!time24hr"
				type="button"
				class="jalali-calendar__ampm"
				:disabled="disabled"
				@click="toggleAmpm"
			>
				{{ ampmLabel }}
			</button>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import {
	formatJalaliDate,
	getJalaliMonthLength,
	instantToJalali,
	jalaliToGregorian,
	jalaliToInstant,
	jalaliWeekday,
	toPersianDigits,
	type JalaliDate,
	type JalaliDateTime,
} from '@/helpers/time/jalali'

const props = withDefaults(defineProps<{
	modelValue: Date | Date[] | null,
	mode?: 'single' | 'range',
	// User timezone the selection is constructed in. Required: the caller owns
	// timezone resolution, this component never assumes UTC or browser zone.
	timeZone: string,
	enableTime?: boolean,
	time24hr?: boolean,
	defaultHour?: number | null,
	defaultMinute?: number | null,
	minDate?: Date | null,
	maxDate?: Date | null,
	disabled?: boolean,
}>(), {
	mode: 'single',
	enableTime: false,
	time24hr: true,
	defaultHour: null,
	defaultMinute: null,
	minDate: null,
	maxDate: null,
	disabled: false,
})

const emit = defineEmits<{
	'update:modelValue': [value: Date | Date[] | null],
}>()

const {t} = useI18n({useScope: 'global'})

// Saturday-first weekday abbreviations. Same set flatpickr ships for fa
// (its weekdays.shorthand, rotated to start Saturday): short enough for the
// seven narrow columns, unlike the full Intl short names.
const WEEKDAY_NAMES = ['شنبه', 'یک', 'دو', 'سه', 'چهار', 'پنج', 'جمعه']

function todayInZone(): JalaliDateTime {
	return instantToJalali(new Date(), props.timeZone) ?? {year: 1405, month: 1, day: 1, hours: 0, minutes: 0}
}

function selectedInstants(): Date[] {
	if (props.modelValue === null) {
		return []
	}
	return Array.isArray(props.modelValue) ? props.modelValue : [props.modelValue]
}

// Viewed month, initialized from the selection or today.
function initialViewed(): JalaliDate {
	const [first] = selectedInstants()
	const parts = first ? instantToJalali(first, props.timeZone) : null
	const today = todayInZone()
	return {year: parts?.year ?? today.year, month: parts?.month ?? today.month, day: 1}
}

const viewed = ref<JalaliDate>(initialViewed())

function stepMonth(delta: number) {
	let {year, month} = viewed.value
	month += delta
	if (month < 1) {
		month = 12
		year -= 1
	} else if (month > 12) {
		month = 1
		year += 1
	}
	viewed.value = {year, month, day: 1}
}

function stepYear(delta: number) {
	viewed.value = {year: viewed.value.year + delta, month: viewed.value.month, day: 1}
}

// Follow external changes (e.g. shortcuts) when they leave the viewed month.
watch(() => props.modelValue, (modelValue) => {
	const first = Array.isArray(modelValue) ? modelValue[0] : modelValue
	const parts = first ? instantToJalali(first, props.timeZone) : null
	if (parts && (parts.year !== viewed.value.year || parts.month !== viewed.value.month)) {
		viewed.value = {year: parts.year, month: parts.month, day: 1}
	}
	// Drop a pending range start the parent never adopted (e.g. it ignored the
	// single-click emission), so a later click cannot complete a stale range.
	const pending = pendingStart.value
	if (pending !== null && !selectedJalali().some((s) => compareJalali(s, pending) === 0)) {
		pendingStart.value = null
	}
	syncTimeFromModel()
})

// Month title and weekday names via the shared Intl helpers.
const monthTitle = computed(() => {
	const gregorian = jalaliToGregorian({year: viewed.value.year, month: viewed.value.month, day: 1})
	if (gregorian === null) {
		return ''
	}
	const representative = new Date(Date.UTC(gregorian.year, gregorian.month - 1, gregorian.day, 12))
	return formatJalaliDate(representative, {timeZone: 'UTC', month: 'long', year: 'numeric'})
})

const weekdayNames = computed(() => WEEKDAY_NAMES)

interface DayCell {
	day: number
	label: string
	ariaLabel: string
	selected: boolean
	today: boolean
	rangeStart: boolean
	rangeEnd: boolean
	inRange: boolean
	hovered: boolean
	disabled: boolean
}

function compareJalali(a: JalaliDate, b: JalaliDate): number {
	return a.year - b.year || a.month - b.month || a.day - b.day
}

function selectedJalali(): JalaliDate[] {
	const dates: JalaliDate[] = []
	for (const instant of selectedInstants()) {
		const parts = instantToJalali(instant, props.timeZone)
		if (parts !== null) {
			dates.push(parts)
		}
	}
	return dates.sort(compareJalali)
}

const minJalali = computed(() => props.minDate ? instantToJalali(props.minDate, props.timeZone) : null)
const maxJalali = computed(() => props.maxDate ? instantToJalali(props.maxDate, props.timeZone) : null)
const todayJalali = computed(() => todayInZone())

const leadingBlanks = computed(() => {
	// jalaliWeekday follows JS convention (0 = Sunday); Saturday-first grid.
	const first = jalaliWeekday({year: viewed.value.year, month: viewed.value.month, day: 1}) ?? 6
	return (first - 6 + 7) % 7
})

const hoveredDay = ref<number | null>(null)
// Range start armed by the first click while the parent has not adopted it
// (parents mirror flatpickr and ignore single-click emissions).
const pendingStart = ref<JalaliDate | null>(null)

const cells = computed<DayCell[]>(() => {
	const length = getJalaliMonthLength(viewed.value.year, viewed.value.month)
	const selected = selectedJalali()
	const [rangeStart, rangeEnd] = props.mode === 'range' && selected.length === 2
		? [selected[0], selected[1]]
		: [null, null]
	const anchor = pendingStart.value ?? (selected.length === 1 ? selected[0] : null)
	const rangeComplete = selected.length === 2 && pendingStart.value === null
	const hoverEnd = props.mode === 'range' && anchor !== null && !rangeComplete && hoveredDay.value !== null
		? {year: viewed.value.year, month: viewed.value.month, day: hoveredDay.value}
		: null

	return Array.from({length}, (_, index) => {
		const day = index + 1
		const current = {year: viewed.value.year, month: viewed.value.month, day}
		const isSelected = selected.some((s) => compareJalali(s, current) === 0)
		const inRange = rangeStart !== null && rangeEnd !== null &&
			compareJalali(current, rangeStart) > 0 && compareJalali(current, rangeEnd) < 0
		const hovered = hoverEnd !== null && anchor !== null &&
			((compareJalali(current, anchor) > 0 && compareJalali(current, hoverEnd) <= 0) ||
				(compareJalali(current, hoverEnd) >= 0 && compareJalali(current, anchor) < 0))
		const disabled = (minJalali.value !== null && compareJalali(current, minJalali.value) < 0) ||
			(maxJalali.value !== null && compareJalali(current, maxJalali.value) > 0)

		return {
			day,
			label: toPersianDigits(day),
			ariaLabel: formatJalaliDate(
				jalaliToInstant({...current, hours: 12, minutes: 0}, props.timeZone),
				{timeZone: props.timeZone, weekday: 'long', day: 'numeric', month: 'long', year: 'numeric'},
			),
			selected: props.mode === 'single' ? isSelected : isSelected || hovered,
			today: compareJalali(current, todayJalali.value) === 0,
			rangeStart: rangeStart !== null && compareJalali(current, rangeStart) === 0,
			rangeEnd: rangeEnd !== null && compareJalali(current, rangeEnd) === 0,
			inRange,
			hovered,
			disabled,
		}
	})
})

// Time state: wall-clock in the user timezone. Initialized from the selection
// so picking another day keeps the time, like flatpickr does.
const hours = ref(0)
const minutes = ref(0)

function currentWallTime(): {hours: number, minutes: number} {
	const now = todayInZone()
	if (props.defaultHour !== null && props.defaultMinute !== null) {
		return {hours: props.defaultHour, minutes: props.defaultMinute}
	}
	return {hours: now.hours, minutes: now.minutes}
}

function syncTimeFromModel() {
	const [first] = selectedInstants()
	const parts = first ? instantToJalali(first, props.timeZone) : null
	const fallback = currentWallTime()
	hours.value = parts?.hours ?? fallback.hours
	minutes.value = parts?.minutes ?? fallback.minutes
}

syncTimeFromModel()

const hoursModel = computed({
	get: () => props.time24hr ? hours.value : ((hours.value + 11) % 12) + 1,
	set: (value: number) => {
		const parsed = Math.trunc(Number(value))
		if (Number.isNaN(parsed)) {
			return
		}
		hours.value = props.time24hr
			? Math.min(23, Math.max(0, parsed))
			: (Math.min(12, Math.max(1, parsed)) % 12) + (hours.value >= 12 ? 12 : 0)
	},
})

const minutesModel = computed({
	get: () => minutes.value,
	set: (value: number) => {
		const parsed = Math.trunc(Number(value))
		if (Number.isNaN(parsed)) {
			return
		}
		minutes.value = Math.min(59, Math.max(0, parsed))
	},
})

const ampmLabel = computed(() => hours.value >= 12 ? t('input.datepicker.pm') : t('input.datepicker.am'))

function toggleAmpm() {
	hours.value = (hours.value + 12) % 24
	emitWithTime()
}

function emitSingle(day: number) {
	const instant = jalaliToInstant(
		{year: viewed.value.year, month: viewed.value.month, day, hours: hours.value, minutes: minutes.value},
		props.timeZone,
	)
	if (instant !== null) {
		emit('update:modelValue', instant)
	}
}

function emitWithTime() {
	if (props.mode === 'range') {
		return
	}
	const selected = selectedJalali()
	if (selected.length === 0) {
		return
	}
	emitSingle(selected[0].day)
}

function selectDay(day: number) {
	if (props.mode === 'range') {
		selectRangeDay(day)
		return
	}
	emitSingle(day)
}

function selectRangeDay(day: number) {
	const current = {year: viewed.value.year, month: viewed.value.month, day}
	const selected = selectedJalali()
	const anchor = pendingStart.value ?? (selected.length === 1 ? selected[0] : null)
	if (anchor === null) {
		// First click arms the range; parents mirror flatpickr and only adopt
		// completed pairs, so the grid tracks the pending start itself.
		pendingStart.value = current
		const instant = jalaliToInstant({...current, hours: 0, minutes: 0}, props.timeZone)
		if (instant !== null) {
			hoveredDay.value = null
			emit('update:modelValue', [instant])
		}
		return
	}

	pendingStart.value = null
	const [from, to] = compareJalali(anchor, current) <= 0 ? [anchor, current] : [current, anchor]
	const fromInstant = jalaliToInstant({...from, hours: 0, minutes: 0}, props.timeZone)
	const toInstant = jalaliToInstant({...to, hours: 23, minutes: 59}, props.timeZone)
	if (fromInstant !== null && toInstant !== null) {
		hoveredDay.value = null
		emit('update:modelValue', [fromInstant, toInstant])
	}
}

function hoverDay(day: number) {
	hoveredDay.value = day
}

// Buttons are natively keyboard accessible; arrow keys move across the grid.
// In RTL the visual left points forward in time.
const dayButtons = ref<HTMLButtonElement[]>([])

function dayIndex(day: number): number {
	return day - 1
}

function focusDay(day: number) {
	const length = getJalaliMonthLength(viewed.value.year, viewed.value.month)
	if (day < 1 || day > length) {
		return
	}
	dayButtons.value[dayIndex(day)]?.focus()
}

function onDayKeydown(event: KeyboardEvent, day: number) {
	const rtlForward = document.dir === 'rtl' ? 1 : -1
	switch (event.key) {
		case 'ArrowLeft':
			event.preventDefault()
			focusDay(day + rtlForward)
			break
		case 'ArrowRight':
			event.preventDefault()
			focusDay(day - rtlForward)
			break
		case 'ArrowUp':
			event.preventDefault()
			focusDay(day - 7)
			break
		case 'ArrowDown':
			event.preventDefault()
			focusDay(day + 7)
			break
	}
}
</script>

<style lang="scss" scoped>
.jalali-calendar {
	inline-size: 100%;
	color: var(--text);
	user-select: none;
}

.jalali-calendar__nav {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-block-end: .5rem;
}

.jalali-calendar__title {
	font-weight: $weight-bold;
	font-size: 1rem;
}

.jalali-calendar__nav-button {
	border: 0;
	background: transparent;
	color: var(--text-light);
	font-size: 1.25rem;
	line-height: 1;
	padding: .25rem .5rem;
	border-radius: $radius;
	cursor: pointer;

	&:hover:not(:disabled) {
		background: var(--grey-100);
		color: var(--text);
	}

	&:disabled {
		opacity: .4;
		cursor: default;
	}
}

.jalali-calendar__weekdays,
.jalali-calendar__days {
	display: grid;
	grid-template-columns: repeat(7, 1fr);
	text-align: center;
}

.jalali-calendar__weekday {
	font-size: .8rem;
	color: var(--text-light);
	padding-block: .25rem;
}

.jalali-calendar__day {
	aspect-ratio: 1;
	border: 0;
	border-radius: $radius;
	background: transparent;
	color: var(--text);
	cursor: pointer;

	&:hover:not(:disabled),
	&:focus-visible {
		background: var(--grey-100);
	}

	&:disabled {
		opacity: .35;
		cursor: default;
	}

	&.is-today {
		border: 1px solid var(--primary);
	}

	&.is-selected,
	&.is-range-start,
	&.is-range-end {
		background: var(--primary);
		color: var(--white);
	}

	&.is-in-range,
	&.is-hovered {
		background: var(--grey-100);
		border-radius: 0;
	}
}

.jalali-calendar__time {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: .25rem;
	margin-block-start: .75rem;
}

.jalali-calendar__time-input {
	inline-size: 3.5rem;
	text-align: center;
}

.jalali-calendar__time-separator {
	font-weight: $weight-bold;
}

.jalali-calendar__ampm {
	border: 1px solid var(--input-border-color);
	border-radius: $radius;
	background: transparent;
	color: var(--text);
	padding: .25rem .5rem;
	cursor: pointer;

	&:hover:not(:disabled) {
		border-color: var(--input-hover-border-color);
	}
}
</style>
