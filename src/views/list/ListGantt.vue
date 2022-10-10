<template>
	<ListWrapper class="list-gantt" :list-id="props.listId" viewName="gantt">
		<template #header>
			<card>
				<div class="gantt-options">
					<div class="field">
						<label class="label" for="range">{{ $t('list.gantt.range') }}</label>
						<div class="control">
							<Foo
								ref="flatPickerEl"
								:config="flatPickerConfig"
								class="input"
								id="range"
								:placeholder="$t('list.gantt.range')"
								v-model="flatPickerDateRange"
							/>
						</div>
					</div>
					<fancycheckbox class="is-block" v-model="showTasksWithoutDates">
						{{ $t('list.gantt.showTasksWithoutDates') }}
					</fancycheckbox>
				</div>
			</card>
		</template>

		<template #default>
			<div class="gantt-chart-container">
				<card :padding="false" class="has-overflow">
					<pre>{{dateRange}}</pre>
					<pre>{{new Date(dateRange.dateFrom).toLocaleDateString()}}</pre>
					<pre>{{new Date(dateRange.dateTo).toLocaleDateString()}}</pre>
					<!-- <gantt-chart
						v-if="false"
						:date-range="dateRange"
						:list-id="props.listId"
						:show-tasks-without-dates="showTasksWithoutDates"
					/> -->

				</card>
			</div>
		</template>
	</ListWrapper>
</template>

<script setup lang="ts">
import {computed, ref, type PropType} from 'vue'
import Foo from '@/components/misc/flatpickr/Flatpickr.vue'
// import type FlatPickr from 'vue-flatpickr-component'
import {useI18n} from 'vue-i18n'
import {format} from 'date-fns'
import {useRoute, useRouter} from 'vue-router'

import {useAuthStore} from '@/stores/auth'

import ListWrapper from './ListWrapper.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'

import {createAsyncComponent} from '@/helpers/createAsyncComponent'

type DateKebab = `${string}-${string}-${string}`

const GanttChart = createAsyncComponent(() => import('@/components/tasks/gantt-chart.vue'))

const props = defineProps({
	listId: {
		type: Number,
		required: true,
	},
	dateFrom: {
		type: String as PropType<DateKebab>,
	},
	dateTo: {
		type: String as PropType<DateKebab>,
	},
	showTasksWithoutDates: {
		type: Boolean,
		default: false,
	},
})

const router = useRouter()
const route = useRoute()

const showTasksWithoutDates = computed({
	get: () => props.showTasksWithoutDates,
	set: (value) => router.push({ query: {
		...route.query,
		showTasksWithoutDates: String(value),
	}}),
})

function parseKebabDate(kebabDate: DateKebab | undefined, fallback: () => Date): Date {
	try {

		if (!kebabDate) {
			throw new Error('No value')
		}
		const dateValues = kebabDate.split('-')
		const [, monthString, dateString] = dateValues
		const [year, month, date] = dateValues.map(val => Number(val))
		const dateValuesAreValid = (
			!Number.isNaN(year) &&
			monthString.length >= 1 && monthString.length <= 2 &&
			!Number.isNaN(month) &&
			month >= 1 && month <= 12 &&
			dateString.length >= 1 && dateString.length <= 31 &&
			!Number.isNaN(date) &&
			date >= 1 && date <= 31
		)
		if (!dateValuesAreValid) {
			throw new Error('Invalid date values')
		}
		return new Date(year, month, date)
	} catch(e) {
		// ignore nonsense route queries
		return fallback()
	}
}

const DEFAULT_DATEFROM_DAY_OFFSET = 0
// const DEFAULT_DATEFROM_DAY_OFFSET = -15
const DEFAULT_DATETO_DAY_OFFSET = +55
// const DEFAULT_DATETO_DAY_OFFSET = +55

const now = new Date()

function getDefaultDateFrom() {
	return new Date(now.getFullYear(), now.getMonth(), now.getDate() + DEFAULT_DATEFROM_DAY_OFFSET)
}

function getDefaultDateTo() {
	return new Date(now.getFullYear(), now.getMonth(), now.getDate() + DEFAULT_DATETO_DAY_OFFSET)
}

let isChangingRoute = ref<ReturnType<typeof router.push> | false>(false)

const count = ref(0)

const dateRange = computed<{
	dateFrom: string
	dateTo: string
}>({
	get: () => ({
		dateFrom: parseKebabDate(props.dateFrom, getDefaultDateFrom).toISOString(),
		dateTo: parseKebabDate(props.dateTo, getDefaultDateTo).toISOString(),
	}),
	async set(range: {
		dateFrom: string
		dateTo: string
	} | null) {
		if (range === null) {
			const query = {...route.query}
			delete query?.dateFrom
			delete query?.dateTo
			console.log('set range to null. query is: ', query)
			router.push(query)
			return
		}
		const {
			dateFrom,
			dateTo,
		} = range
		count.value = count.value + 1
		if (count.value >= 4) {
			console.log('triggered ', count, ' times, stopping.')
			return
		}
		if (isChangingRoute.value !== false) {
			console.log('called again while changing route')
			await isChangingRoute.value
			console.log('changing route finished, continuing...')
		}

		const queryDateFrom = format(new Date(dateFrom || getDefaultDateFrom()), 'yyyy-LL-dd')
		const queryDateTo = format(new Date(dateTo || getDefaultDateTo()), 'yyyy-LL-dd')

		console.log(dateFrom, 'dateFrom')
		console.log(dateRange.value.dateFrom, 'dateRange.value.dateFrom')
		console.log(dateTo, 'dateTo')
		console.log(dateRange.value.dateTo, 'dateRange.value.dateTo')

		if (queryDateFrom === route.query.dateFrom || queryDateTo === route.query.dateTo) {
			console.log('is same date')
			// only set if the value has changed
			return
		}
		console.log('change url to', {
			query: {
				...route.query,
				dateFrom: format(new Date(dateFrom), 'yyyy-LL-dd'),
				dateTo: format(new Date(dateTo), 'yyyy-LL-dd'),
			}
		})
		isChangingRoute.value = router.push({
			query: {
				...route.query,
				dateFrom: format(new Date(dateFrom), 'yyyy-LL-dd'),
				dateTo: format(new Date(dateTo), 'yyyy-LL-dd'),
			}
		})
	},
})

const initialDateRange = [dateRange.value.dateFrom, dateRange.value.dateTo]

function getCurrentDateRangeFromFlatpicker() {
	return flatPickerEl.value.fp.selectedDates.map((date: Date) => date?.toISOString())
}

const flatPickerEl = ref<typeof FlatPickr | null>(null)
const flatPickerDateRange = computed({
	get: () => ([
		dateRange.value.dateFrom,
		dateRange.value.dateTo
	]),
	set(newVal) {
	// set([dateFrom, dateTo]) {
		// newVal from event does only contain the wrong format
		console.log(newVal)
		const [dateFrom, dateTo] = newVal
		// const [dateFrom, dateTo] = getCurrentDateRangeFromFlatpicker()
		
		if (
			// only set after whole range has been selected
			!dateTo ||
			// only set if the value has changed
			dateRange.value.dateFrom === dateFrom &&
			dateRange.value.dateTo === dateTo
		) {
			return
		}
		// dateRange.value = {dateFrom, dateTo}
	}
})

const ISO_DATE_FORMAT = "YYYY-MM-DDTHH:mm:ssZ[Z]"

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatShort'),
	altInput: true,
	// dateFornat: ISO_DATE_FORMAT,
	// dateFormat: 'Y-m-d',
	defaultDate: initialDateRange,
	enableTime: false,
	mode: 'range',
	locale: {
		firstDayOfWeek: authStore.settings.weekStart,
	},
}))
</script>

<style lang="scss">
.gantt-chart-container {
	padding-bottom: 1rem;
}

.gantt-options {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-bottom: 1rem;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}

	.field {
		margin-bottom: 0;
		width: 33%;

		&:not(:last-child) {
			padding-right: .5rem;
		}

		@media screen and (max-width: $tablet) {
			width: 100%;
			max-width: 100%;
			margin-top: .5rem;
			padding-right: 0 !important;
		}

		&, .input {
			font-size: .8rem;
		}

		.select, .select select {
			height: auto;
			width: 100%;
			font-size: .8rem;
		}


		.label {
			font-size: .9rem;
			padding-left: .4rem;
		}
	}
}

// vue-draggable overwrites
.vdr.active::before {
	display: none;
}

.link-share-view:not(.has-background) .card.gantt-options {
	border: none;
	box-shadow: none;

	.card-content {
		padding: .5rem;
	}
}
</style>