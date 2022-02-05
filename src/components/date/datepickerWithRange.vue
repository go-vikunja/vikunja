<template>
	<div class="datepicker-with-range-container">
		<popup>
			<template #trigger="{toggle}">
				<slot name="trigger" :toggle="toggle">
					<x-button @click.prevent.stop="toggle()" type="secondary" :shadow="false" class="mb-2">
						{{ $t('task.show.select') }}
					</x-button>
				</slot>
			</template>
			<template #content="{isOpen}">
				<div class="datepicker-with-range" :class="{'is-open': isOpen}">
					<div class="selections">
						<button @click="setDateRange(datesToday)" :class="{'is-active': dateRange === datesToday}">
							{{ $t('task.show.today') }}
						</button>
						<button @click="setDateRange(datesThisWeek)"
								:class="{'is-active': dateRange === datesThisWeek}">
							{{ $t('task.show.thisWeek') }}
						</button>
						<button @click="setDateRange(datesNextWeek)"
								:class="{'is-active': dateRange === datesNextWeek}">
							{{ $t('task.show.nextWeek') }}
						</button>
						<button @click="setDateRange(datesNext7Days)"
								:class="{'is-active': dateRange === datesNext7Days}">
							{{ $t('task.show.next7Days') }}
						</button>
						<button @click="setDateRange(datesThisMonth)"
								:class="{'is-active': dateRange === datesThisMonth}">
							{{ $t('task.show.thisMonth') }}
						</button>
						<button @click="setDateRange(datesNextMonth)"
								:class="{'is-active': dateRange === datesNextMonth}">
							{{ $t('task.show.nextMonth') }}
						</button>
						<button @click="setDateRange(datesNext30Days)"
								:class="{'is-active': dateRange === datesNext30Days}">
							{{ $t('task.show.next30Days') }}
						</button>
						<button @click="setDateRange('')" :class="{'is-active': customRangeActive}">
							{{ $t('misc.custom') }}
						</button>
					</div>
					<div class="flatpickr-container">
						<flat-pickr
							:config="flatPickerConfig"
							v-model="dateRange"
						/>
					</div>
				</div>
			</template>
		</popup>
	</div>
</template>

<script lang="ts" setup>
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {store} from '@/store'
import {format} from 'date-fns'
import Popup from '@/components/misc/popup'

const {t} = useI18n()

const emit = defineEmits(['dateChanged'])

// FIXME: This seems to always contain the default value - that breaks the picker
const weekStart = computed<number>(() => store.state.auth.settings.weekStart ?? 0)
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: false,
	inline: true,
	mode: 'range',
	locale: {
		firstDayOf7Days: weekStart.value,
	},
}))

const dateRange = ref('')

watch(
	() => dateRange.value,
	(newVal: string | null) => {
		if (newVal === null) {
			return
		}

		const [fromDate, toDate] = newVal.split(' to ')

		if (typeof fromDate === 'undefined' || typeof toDate === 'undefined') {
			return
		}

		emit('dateChanged', {
			dateFrom: new Date(fromDate),
			dateTo: new Date(toDate),
		})
	},
)

function formatDate(date: Date): string {
	return format(date, 'yyyy-MM-dd HH:mm')
}

function startOfDay(date: Date): Date {
	date.setHours(0)
	date.setMinutes(0)

	return date
}

function endOfDay(date: Date): Date {
	date.setHours(23)
	date.setMinutes(59)

	return date
}

const datesToday = computed<string>(() => {
	const startDate = startOfDay(new Date())
	const endDate = endOfDay(new Date())

	return `${formatDate(startDate)} to ${formatDate(endDate)}`
})

function thisWeek() {
	const startDate = startOfDay(new Date())
	const first = startDate.getDate() - startDate.getDay() + weekStart.value
	startDate.setDate(first)
	const endDate = endOfDay(new Date((new Date(startDate).setDate(first + 6))))

	return {
		startDate,
		endDate,
	}
}

const datesThisWeek = computed<string>(() => {
	const {startDate, endDate} = thisWeek()

	return `${formatDate(startDate)} to ${formatDate(endDate)}`
})

const datesNextWeek = computed<string>(() => {
	const {startDate, endDate} = thisWeek()
	startDate.setDate(startDate.getDate() + 7)
	endDate.setDate(endDate.getDate() + 7)

	return `${formatDate(startDate)} to ${formatDate(endDate)}`
})

const datesNext7Days = computed<string>(() => {
	const startDate = startOfDay(new Date())
	const endDate = endOfDay(new Date((new Date()).getTime() + 7 * 24 * 60 * 60 * 1000))
	return `${formatDate(startDate)} to ${formatDate(endDate)}`
})

function thisMonth() {
	const startDate = startOfDay(new Date())
	startDate.setDate(1)
	const endDate = endOfDay(new Date((new Date()).getFullYear(), (new Date()).getMonth() + 1, 0))

	return {
		startDate,
		endDate,
	}
}

const datesThisMonth = computed<string>(() => {
	const {startDate, endDate} = thisMonth()

	return `${formatDate(startDate)} to ${formatDate(endDate)}`
})

const datesNextMonth = computed<string>(() => {
	const {startDate, endDate} = thisMonth()

	startDate.setMonth(startDate.getMonth() + 1)
	endDate.setMonth(endDate.getMonth() + 1)

	return `${formatDate(startDate)} to ${formatDate(endDate)}`
})

const datesNext30Days = computed<string>(() => {
	const startDate = startOfDay(new Date())
	const endDate = endOfDay(new Date((new Date()).setMonth((new Date()).getMonth() + 1)))

	return `${formatDate(startDate)} to ${formatDate(endDate)}`
})

function setDateRange(range: string) {
	dateRange.value = range
}

const customRangeActive = computed<Boolean>(() => {
	return dateRange.value !== datesToday.value &&
		dateRange.value !== datesThisWeek.value &&
		dateRange.value !== datesNextWeek.value &&
		dateRange.value !== datesNext7Days.value &&
		dateRange.value !== datesThisMonth.value &&
		dateRange.value !== datesNextMonth.value &&
		dateRange.value !== datesNext30Days.value
})
</script>

<style lang="scss" scoped>
.datepicker-with-range-container {
	position: relative;

	:deep(.popup) {
		z-index: 10;
		margin-top: 1rem;
		border-radius: $radius;
		border: 1px solid var(--grey-200);
		background-color: var(--white);
		box-shadow: $shadow;

		&.is-open {
			width: 500px;
			height: 320px;
		}
	}
}

.datepicker-with-range {
	display: flex;
	width: 100%;
	height: 100%;
	position: absolute;

	:deep(.flatpickr-calendar) {
		margin: 0 auto 8px;
		box-shadow: none;
	}
}

.flatpickr-container {
	width: 70%;
	border-left: 1px solid var(--grey-200);

	:deep(input.input) {
		display: none;
	}
}

.selections {
	width: 30%;
	display: flex;
	flex-direction: column;
	padding-top: .5rem;

	button {
		display: block;
		width: 100%;
		text-align: left;
		padding: .5rem 1rem;
		transition: $transition;
		font-size: .9rem;
		color: var(--text);
		background: transparent;
		border: 0;
		cursor: pointer;

		&.is-active {
			color: var(--primary);
		}

		&:hover, &.is-active {
			background-color: var(--grey-100);
		}
	}
}
</style>
