<template>
	<div
		class="datepicker-inline"
		:class="`datepicker-inline--${shortcutsLayout}`"
	>
		<DateShortcuts
			v-if="showShortcuts && shortcutsLayout === 'chips'"
			layout="chips"
			:active="date"
			@select="setShortcut"
		/>

		<div class="datepicker-inline__body">
			<DateShortcuts
				v-if="showShortcuts && shortcutsLayout === 'sidebar'"
				layout="list"
				:active="date"
				@select="setShortcut"
			/>
			<CalendarMonth
				class="datepicker-inline__calendar"
				:selected="date"
				:min-date="minDate"
				:large="large"
				@pick="setDay"
			/>
		</div>

		<TimeControl
			:hours="date?.getHours() ?? defaultTime.hours"
			:minutes="date?.getMinutes() ?? defaultTime.minutes"
			@update="setTime"
		/>
	</div>
</template>

<script lang="ts" setup>
import {computed, ref, toRef, watch} from 'vue'

import CalendarMonth from '@/components/input/datepicker/CalendarMonth.vue'
import DateShortcuts from '@/components/input/datepicker/DateShortcuts.vue'
import TimeControl from '@/components/input/datepicker/TimeControl.vue'

import {createDateFromString} from '@/helpers/time/createDateFromString'
import {getDateWithTime, parseUserDefaultTime} from '@/helpers/time/getDateWithTime'
import {useAuthStore} from '@/stores/auth'

const props = withDefaults(defineProps<{
	modelValue: Date | null | string
	showShortcuts?: boolean
	// Chips fit narrow containers (mobile sheet, reminder popup); the sidebar needs the full popup width.
	shortcutsLayout?: 'sidebar' | 'chips'
	minDate?: Date | null
	large?: boolean
}>(), {
	showShortcuts: true,
	shortcutsLayout: 'sidebar',
	minDate: null,
	large: false,
})

const emit = defineEmits<{
	'update:modelValue': [Date | null],
}>()

const date = ref<Date | null>(null)

watch(
	toRef(props, 'modelValue'),
	(value) => {
		date.value = value === null || value === '' ? null : createDateFromString(value)
	},
	{immediate: true},
)

const authStore = useAuthStore()
// Noon when the user has no default due time configured.
const defaultTime = computed(() => parseUserDefaultTime(authStore.settings.frontendSettings.defaultDueTime) ?? {hours: 12, minutes: 0})

function update(value: Date | null) {
	// The calendar only compares days, so a picked day can still carry a time before minDate.
	const clamped = value !== null && props.minDate && value < props.minDate
		? new Date(props.minDate)
		: value
	date.value = clamped
	emit('update:modelValue', clamped)
}

function withCurrentTime(day: Date, fallback: Date): Date {
	const result = new Date(day)
	if (date.value) {
		result.setHours(date.value.getHours(), date.value.getMinutes(), 0, 0)
		return result
	}
	return fallback
}

function setDay(day: Date) {
	const fallback = new Date(day)
	fallback.setHours(defaultTime.value.hours, defaultTime.value.minutes, 0, 0)
	update(withCurrentTime(day, fallback))
}

// Shortcuts pick the next sensible hour ("later today"), unlike a plain calendar click.
function setShortcut(day: Date) {
	update(withCurrentTime(day, getDateWithTime(day)))
}

function setTime({hours, minutes}: {hours: number, minutes: number}) {
	const result = new Date(date.value ?? new Date())
	result.setHours(hours, minutes, 0, 0)
	update(result)
}
</script>

<style lang="scss" scoped>
.datepicker-inline {
	display: flex;
	flex-direction: column;
	inline-size: 100%;
}

.datepicker-inline__body {
	display: flex;
}

.datepicker-inline__calendar {
	flex: 1;
	min-inline-size: 0;
	padding: .75rem 1rem 1rem;
}
</style>
