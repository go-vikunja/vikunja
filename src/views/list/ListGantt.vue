<template>
	<ListWrapper class="list-gantt" :list-id="filters.listId" viewName="gantt">
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
					<fancycheckbox class="is-block" v-model="filters.showTasksWithoutDates">
						{{ $t('list.gantt.showTasksWithoutDates') }}
					</fancycheckbox>
				</div>
			</card>
		</template>

		<template #default>
			<div class="gantt-chart-container">
				<card :padding="false" class="has-overflow">
					<pre>{{dateRange}}</pre>
					<pre>{{new Date(dateRange.dateFrom).toISOString()}}</pre>
					<pre>{{new Date(dateRange.dateTo).toISOString()}}</pre>
					<gantt-chart
						:list-id="filters.listId"
						:date-from="filters.dateFrom"
						:date-to="filters.dateTo"
						:show-tasks-without-dates="showTasksWithoutDates"
					/>
				</card>
			</div>
		</template>
	</ListWrapper>
</template>

<script setup lang="ts">
import {computed, reactive, ref, watch, type PropType} from 'vue'
import Foo from '@/components/misc/flatpickr/Flatpickr.vue'
// import type FlatPickr from 'vue-flatpickr-component'
import type Flatpickr from 'flatpickr'
import {useI18n} from 'vue-i18n'
import {format} from 'date-fns'
import {useRoute, useRouter, type LocationQuery, type RouteLocationNormalized, type RouteLocationRaw} from 'vue-router'

import {useAuthStore} from '@/stores/auth'

import ListWrapper from './ListWrapper.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'

import {createAsyncComponent} from '@/helpers/createAsyncComponent'
import GanttChart from '@/components/tasks/gantt-chart.vue'
import type { IList } from '@/modelTypes/IList'

export type DateKebab = `${string}-${string}-${string}`
export type DateISO = string
export type DateRange = {
	dateFrom: string
	dateTo: string
}

export interface GanttParams {
	listId: IList['id']
	dateFrom: DateKebab
	dateTo: DateKebab
	showTasksWithoutDates: boolean
	route: RouteLocationNormalized,
}

export interface GanttFilter {
	listId: IList['id']
	dateFrom: DateISO
	dateTo: DateISO
	showTasksWithoutDates: boolean
}

type Options = Flatpickr.Options.Options

// const GanttChart = createAsyncComponent(() => import('@/components/tasks/gantt-chart.vue'))

const props = defineProps<GanttParams>()

const router = useRouter()
const route = useRoute()

function parseDateProp(kebabDate: DateKebab | undefined): string | undefined {
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
		return new Date(year, month, date).toISOString()
	} catch(e) {
		// ignore nonsense route queries
		return
	}
}

function parseBooleanProp(booleanProp: string) {
	return (booleanProp === 'false' || booleanProp === '0')
		? false
		:	Boolean(booleanProp)
}

const DATE_FORMAT_KEBAB = 'yyyy-LL-dd'
function isoToKebabDate(isoDate: DateISO) {
	return format(new Date(isoDate), DATE_FORMAT_KEBAB) as DateKebab
}

const DEFAULT_SHOW_TASKS_WITHOUT_DATES = false

const DEFAULT_DATEFROM_DAY_OFFSET = -15
const DEFAULT_DATETO_DAY_OFFSET = +55

const now = new Date()

function getDefaultDateFrom() {
	return new Date(now.getFullYear(), now.getMonth(), now.getDate() + DEFAULT_DATEFROM_DAY_OFFSET).toISOString()
}

function getDefaultDateTo() {
	return new Date(now.getFullYear(), now.getMonth(), now.getDate() + DEFAULT_DATETO_DAY_OFFSET).toISOString()
}

function routeToFilter(route: RouteLocationNormalized): GanttFilter {
	console.log('parseDateProp', parseDateProp(route.query.dateTo as DateKebab))
	console.log(parseDateProp(route.query.dateTo as DateKebab))
	return {
		listId: Number(route.params.listId as string),
		dateFrom: parseDateProp(route.query.dateFrom as DateKebab) || getDefaultDateFrom(),
		dateTo: parseDateProp(route.query.dateTo as DateKebab) || getDefaultDateTo(),
		showTasksWithoutDates: parseBooleanProp(route.query.showTasksWithoutDates as string) || DEFAULT_SHOW_TASKS_WITHOUT_DATES,
	}
}

function filterToRoute(filters: GanttFilter): RouteLocationRaw {
	let query: Record<string, string> = {}
	if (
		filters.dateFrom !== getDefaultDateFrom() ||
		filters.dateTo !== getDefaultDateTo()
	) {
		query = {
			dateFrom: isoToKebabDate(filters.dateFrom),
			dateTo: isoToKebabDate(filters.dateTo),
		}
	}

	if (filters.showTasksWithoutDates) {
		query.showTasksWithoutDates = String(filters.showTasksWithoutDates)
	}

	return {
		name: 'list.gantt',
		params: {listId: filters.listId},
		query
	}
}

const filters: GanttFilter = reactive(routeToFilter(route))

watch(() => JSON.parse(JSON.stringify(props.route)) as RouteLocationNormalized, (route, oldRoute) => {
	if (route.name !== oldRoute.name) {
		return
	}
	const filterFullPath = router.resolve(filterToRoute(filters)).fullPath
	if (filterFullPath === route.fullPath) {
		return
	}

	Object.assign(filters, routeToFilter(route))
})

watch(
	filters,
	async () => {
		const newRouteFullPath = router.resolve(filterToRoute(filters)).fullPath
		if (newRouteFullPath !== route.fullPath) {
			await router.push(newRouteFullPath)
		}
	},
	{flush: "post"}
)

const dateRange = computed(() => ({
	dateFrom: filters.dateFrom,
	dateTo: filters.dateTo,
}))

const flatPickerEl = ref<typeof FlatPickr | null>(null)
const flatPickerDateRange = computed({
	get: () => ([
		filters.dateFrom,
		filters.dateTo
	]),
	set(newVal) {
		const [dateFrom, dateTo] = newVal.map((date) => date?.toISOString())
		
		// only set after whole range has been selected
		if (!dateTo) return

		Object.assign(filters, {dateFrom, dateTo})
	}
})

const ISO_DATE_FORMAT = "YYYY-MM-DDTHH:mm:ssZ[Z]"

const initialDateRange = [filters.dateFrom, filters.dateTo]

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const flatPickerConfig = computed<Options>(() => ({
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