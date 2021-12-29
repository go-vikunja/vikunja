<template>
	<div class="datepicker-with-range">
		<div class="selections">
			<button @click="setDateRange(datesToday)" :class="{'is-active': dateRange === datesToday}">
				{{ $t('task.show.today') }}
			</button>
			<button @click="setDateRange(datesNextWeek)" :class="{'is-active': dateRange === datesNextWeek}">
				{{ $t('task.show.nextWeek') }}
			</button>
			<button @click="setDateRange(datesNextMonth)" :class="{'is-active': dateRange === datesNextMonth}">
				{{ $t('task.show.nextMonth') }}
			</button>
			<button @click="setDateRange('')"  :class="{'is-active': customRangeActive}">
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

<script lang="ts" setup>
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {store} from '@/store'
import {format} from 'date-fns'

const {t} = useI18n()

const emit = defineEmits(['dateChanged'])

const weekStart = computed<number>(() => store.state.auth.settings.weekStart)
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: false,
	inline: true,
	mode: 'range',
	/*locale: {
		// FIXME: This seems to always contain the default value - that breaks the picker
		firstDayOfWeek: weekStart,
	},*/
}))

const dateRange = ref<string>('')

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
	}
)

function formatDate(date: Date): string {
	return format(date, 'yyyy-MM-dd HH:mm')
}

const datesToday = computed<string>(() => {
	const startDate = new Date()
	const endDate = new Date((new Date()).setDate((new Date()).getDate() + 1))
	return `${formatDate(startDate)} to ${formatDate(endDate)}`
})

const datesNextWeek = computed<string>(() => {
	const startDate = new Date()
	const endDate = new Date((new Date()).getTime() + 7 * 24 * 60 * 60 * 1000)
	return `${formatDate(startDate)} to ${formatDate(endDate)}`
})

const datesNextMonth = computed<string>(() => {
	const startDate = new Date()
	const endDate = new Date((new Date()).setMonth((new Date()).getMonth() + 1))
	return `${formatDate(startDate)} to ${formatDate(endDate)}`
})

function setDateRange(range: string) {
	dateRange.value = range
}

const customRangeActive = computed<Boolean>(() => {
	return dateRange.value !== datesToday.value &&
		dateRange.value !== datesNextWeek.value &&
		dateRange.value !== datesNextMonth.value
})
</script>

<style lang="scss" scoped>
.datepicker-with-range {
	border-radius: $radius;
	border: 1px solid var(--grey-200);
	background-color: var(--white);
	box-shadow: $shadow;
	display: flex;
	width: 500px;

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
