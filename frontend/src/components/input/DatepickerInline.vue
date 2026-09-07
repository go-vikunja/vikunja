<template>
	<template v-if="showShortcuts">
		<BaseButton
			v-if="(new Date()).getHours() < 21"
			class="datepicker__quick-select-date"
			@click.stop="setDate('today')"
		>
			<span class="icon"><Icon :icon="['far', 'calendar-alt']" /></span>
			<span class="text">
				<span>{{ $t('input.datepicker.today') }}</span>
				<span class="weekday">{{ getWeekdayFromStringInterval('today') }}</span>
			</span>
		</BaseButton>
		<BaseButton
			class="datepicker__quick-select-date"
			@click.stop="setDate('tomorrow')"
		>
			<span class="icon"><Icon :icon="['far', 'sun']" /></span>
			<span class="text">
				<span>{{ $t('input.datepicker.tomorrow') }}</span>
				<span class="weekday">{{ getWeekdayFromStringInterval('tomorrow') }}</span>
			</span>
		</BaseButton>
		<BaseButton
			class="datepicker__quick-select-date"
			@click.stop="setDate('nextMonday')"
		>
			<span class="icon"><Icon icon="coffee" /></span>
			<span class="text">
				<span>{{ $t('input.datepicker.nextMonday') }}</span>
				<span class="weekday">{{ getWeekdayFromStringInterval('nextMonday') }}</span>
			</span>
		</BaseButton>
		<BaseButton
			v-if="!((new Date()).getDay() === 0 && (new Date()).getHours() >= 21)"
			class="datepicker__quick-select-date"
			@click.stop="setDate('thisWeekend')"
		>
			<span class="icon"><Icon icon="cocktail" /></span>
			<span class="text">
				<span>{{ $t('input.datepicker.thisWeekend') }}</span>
				<span class="weekday">{{ getWeekdayFromStringInterval('thisWeekend') }}</span>
			</span>
		</BaseButton>
		<BaseButton
			class="datepicker__quick-select-date"
			@click.stop="setDate('laterThisWeek')"
		>
			<span class="icon"><Icon icon="chess-knight" /></span>
			<span class="text">
				<span>{{ $t('input.datepicker.laterThisWeek') }}</span>
				<span class="weekday">{{ getWeekdayFromStringInterval('laterThisWeek') }}</span>
			</span>
		</BaseButton>
		<BaseButton
			class="datepicker__quick-select-date"
			@click.stop="setDate('nextWeek')"
		>
			<span class="icon"><Icon icon="forward" /></span>
			<span class="text">
				<span>{{ $t('input.datepicker.nextWeek') }}</span>
				<span class="weekday">{{ getWeekdayFromStringInterval('nextWeek') }}</span>
			</span>
		</BaseButton>
	</template>

	<div class="flatpickr-container">
		<JalaliCalendarGrid
			v-if="isJalali"
			:model-value="date"
			:time-zone="timeZone"
			:enable-time="true"
			:time-24hr="timeFormat === TIME_FORMAT.HOURS_24"
			:default-hour="configuredDueTime?.hours ?? null"
			:default-minute="configuredDueTime?.minutes ?? null"
			@update:modelValue="onGridDate"
		/>
		<flat-pickr
			v-else
			ref="flatPickrRef"
			v-model="flatPickrDate"
			:config="flatPickerConfig"
		/>
	</div>
</template>

<script lang="ts" setup>
import {computed, onBeforeUnmount, onMounted, ref, toRef, watch} from 'vue'
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'

import BaseButton from '@/components/base/BaseButton.vue'
import JalaliCalendarGrid from '@/components/input/JalaliCalendarGrid.vue'

import {formatDate} from '@/helpers/time/formatDate'
import {calculateDayInterval} from '@/helpers/time/calculateDayInterval'
import {createDateFromString} from '@/helpers/time/createDateFromString'
import {getDateWithTime, getDefaultTimeParts, parseUserDefaultTime} from '@/helpers/time/getDateWithTime'
import {
	addJalaliDays,
	instantToJalali,
	jalaliToInstant,
	weekdayInTimezone,
} from '@/helpers/time/jalali'
import {useAuthStore} from '@/stores/auth'
import {useI18n} from 'vue-i18n'
import {useFlatpickrLanguage} from '@/helpers/useFlatpickrLanguage'
import {useJalaliCalendar} from '@/composables/useJalaliCalendar'
import {useTimeFormat} from '@/composables/useTimeFormat'
import {TIME_FORMAT} from '@/constants/timeFormat'

const props = withDefaults(defineProps<{
	modelValue: Date | null | string
	showShortcuts?: boolean
}>(), {
	showShortcuts: true,
})

const emit = defineEmits<{
	'update:modelValue': [Date | null],
}>()

const {t} = useI18n({useScope: 'global'})
const {store: timeFormat} = useTimeFormat()
const {isJalali, timeZone} = useJalaliCalendar()

const date = ref<Date | null>(null)
const changed = ref(false)

const modelValue = toRef(props, 'modelValue')
watch(
	modelValue,
	setDateValue,
	{immediate: true},
)

const flatPickrRef = ref<InstanceType<typeof flatPickr> | null>(null)
const configuredDueTime = computed(() => parseUserDefaultTime(useAuthStore().settings.frontendSettings.defaultDueTime))
const flatPickerConfig = computed(() => {
	const configuredDueTimeValue = configuredDueTime.value

	return {
		altFormat: t('date.altFormatLong'),
		altInput: true,
		dateFormat: 'Y-m-d H:i',
		...(configuredDueTimeValue === null ? {} : {
			defaultHour: configuredDueTimeValue.hours,
			defaultMinute: configuredDueTimeValue.minutes,
		}),
		enableTime: true,
		time_24hr: timeFormat.value === TIME_FORMAT.HOURS_24,
		inline: true,
		locale: useFlatpickrLanguage().value,
	}
})

function formatDateToFlatpickrString(date: Date): string {
	const year = date.getFullYear()
	const month = (date.getMonth() + 1).toString().padStart(2, '0')
	const day = date.getDate().toString().padStart(2, '0')
	const hours = date.getHours().toString().padStart(2, '0')
	const minutes = date.getMinutes().toString().padStart(2, '0')
	
	return `${year}-${month}-${day} ${hours}:${minutes}`
}

// Since flatpickr dates are strings, we need to convert them to native date objects.
// To make that work, we need a separate variable since flatpickr does not have a change event.
const flatPickrDate = computed({
	set(newValue: string | Date | null) {
		if (newValue === null) {
			date.value = null
			return
		}

		if (date.value && formatDateToFlatpickrString(date.value) === newValue) {
			return
		}
		date.value = createDateFromString(newValue)
		updateData()
	},
	get() {
		if (!date.value) {
			return ''
		}
		
		return formatDateToFlatpickrString(date.value)
	},
})

onMounted(() => {
	const inputs = flatPickrRef.value?.$el.parentNode.querySelectorAll('.numInputWrapper > input.numInput')
	inputs?.forEach((i: Element) => {
		i.addEventListener('input', handleFlatpickrInput)
	})
})

onBeforeUnmount(() => {
	const inputs = flatPickrRef.value?.$el.parentNode.querySelectorAll('.numInputWrapper > input.numInput')
	inputs?.forEach((i: Element) => {
		i.removeEventListener('input', handleFlatpickrInput)
	})
})

// Flatpickr only returns a change event when the value in the input it's referring to changes.
// That means it will usually only trigger when the focus is moved out of the input field.
// This is fine most of the time. However, since we're displaying flatpickr in a popup,
// the whole html dom instance might get destroyed, before the change event had a
// chance to fire. In that case, it would not update the date value. To fix
// this, we're now listening on every change and bubble them up as soon
// as they happen.
function handleFlatpickrInput(e: Event) {
	const newDate = new Date(date?.value || 'now')
	const target = e.target as HTMLInputElement
	if (target.classList.contains('flatpickr-minute')) {
		newDate.setMinutes(Number(target.value))
	}
	if (target.classList.contains('flatpickr-hour')) {
		newDate.setHours(Number(target.value))
	}
	if (target.classList.contains('cur-year')) {
		newDate.setFullYear(Number(target.value))
	}
	flatPickrDate.value = newDate
}


function setDateValue(dateString: string | Date | null) {
	if (dateString === null) {
		date.value = null
		return
	}
	date.value = createDateFromString(dateString)
}

function updateData() {
	changed.value = true
	emit('update:modelValue', date.value)
}

function onGridDate(value: Date | Date[] | null) {
	date.value = Array.isArray(value) ? (value[0] ?? null) : value
	updateData()
}

function setDate(dateString: string) {
	if (isJalali.value) {
		setJalaliShortcutDate(dateString)
		return
	}

	const interval = calculateDayInterval(dateString)
	const newDate = new Date()
	newDate.setDate(newDate.getDate() + interval)
	date.value = getDateWithTime(newDate)
	updateData()
}

// Shortcuts keep their existing day-offset semantics; only the "today" anchor
// and the resulting instant move into the user's timezone so the picked date
// round-trips with the Jalali display.
function setJalaliShortcutDate(dateString: string) {
	const tz = timeZone.value
	const now = new Date()
	const nowJalali = instantToJalali(now, tz)
	if (nowJalali === null) {
		return
	}

	const interval = calculateDayInterval(dateString, weekdayInTimezone(now, tz))
	const target = addJalaliDays({year: nowJalali.year, month: nowJalali.month, day: nowJalali.day}, interval)
	// Reinterpret the user-timezone wall clock as local fields so the existing
	// default-time rules (configured due time or nearest hours) apply to what
	// the user actually sees.
	const wallClock = new Date(target.year, target.month - 1, target.day, nowJalali.hours, nowJalali.minutes)
	const {hours, minutes} = getDefaultTimeParts(wallClock)
	const instant = jalaliToInstant({...target, hours, minutes}, tz)
	if (instant === null) {
		return
	}

	date.value = instant
	updateData()
}

function getWeekdayFromStringInterval(dateString: string) {
	const interval = calculateDayInterval(dateString)
	const newDate = new Date()
	newDate.setDate(newDate.getDate() + interval)
	return formatDate(newDate, 'ddd')
}
</script>

<style lang="scss" scoped>
.datepicker__quick-select-date {
	display: flex;
	align-items: center;
	padding: 0 .5rem;
	inline-size: 100%;
	block-size: 2.25rem;
	color: var(--text);
	transition: all $transition;

	&:first-child {
		border-radius: $radius $radius 0 0;
	}

	&:hover {
		background: var(--grey-100);
	}

	.text {
		inline-size: 100%;
		font-size: .85rem;
		display: flex;
		justify-content: space-between;
		padding-inline-end: .25rem;

		.weekday {
			color: var(--text-light);
			text-transform: capitalize;
		}
	}

	.icon {
		inline-size: 2rem;
		text-align: center;
	}
}

.flatpickr-container {
	:deep(.flatpickr-calendar) {
		margin: 0 auto 8px;
		box-shadow: none;
	}

	:deep(.input) {
		border: none;
	}
}
</style>
