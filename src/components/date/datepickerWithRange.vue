<template>
	<div class="datepicker-with-range">
		<div class="selections">
			<a @click="setDatesToToday">Today</a>
			<a @click="setDatesToNextWeek">Next Week</a>
			<a @click="setDatesToNextMonth">Next Month</a>
			<a>Custom</a>
		</div>
		<div class="flatpickr-container">
			<flat-pickr
				:config="flatPickerConfig"
				v-model="dateRange"
			/>
			{{ dateRange }}
		</div>
	</div>
</template>

<script setup>
import flatPickr from 'vue-flatpickr-component'
import 'flatpickr/dist/flatpickr.css'
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {store} from '@/store'
import {format} from 'date-fns'

const {t} = useI18n()

const emit = defineEmits(['dateChanged'])

const weekStart = computed(() => store.state.auth.settings.weekStart)
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatLong'),
	altInput: true,
	dateFormat: 'Y-m-d H:i',
	enableTime: true,
	time_24hr: true,
	inline: true,
	mode: 'range',
	locale: {
		// FIXME: This seems to always contain the default value
		firstDayOfWeek: weekStart,
	},
}))

const dateRange = ref('')

watch(
	() => dateRange.value,
	newVal => {
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

function formatDate(date) {
	return format(date, 'yyyy-MM-dd HH:mm')
}

function setDatesToToday() {
	const startDate = new Date()
	const endDate = new Date((new Date()).setDate((new Date()).getDate() + 1))
	dateRange.value = `${formatDate(startDate)} to ${formatDate(endDate)}`
}

function setDatesToNextWeek() {
	const startDate = new Date()
	const endDate = new Date((new Date()).getTime() + 7 * 24 * 60 * 60 * 1000)
	dateRange.value = `${formatDate(startDate)} to ${formatDate(endDate)}`
}

function setDatesToNextMonth() {
	const startDate = new Date()
	const endDate = new Date((new Date()).setMonth((new Date()).getMonth() + 1))
	dateRange.value = `${formatDate(startDate)} to ${formatDate(endDate)}`
}
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

	a {
		display: block;
		width: 100%;
		text-align: left;
		padding: .5rem 1rem;
		transition: $transition;
		font-size: .9rem;
		color: var(--text);

		&.active {
			color: var(--primary);
		}

		&:hover, &.active {
			background-color: var(--grey-100);
		}
	}
}
</style>
