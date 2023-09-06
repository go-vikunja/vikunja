<template>
	<ProjectWrapper class="project-gantt" :project-id="filters.projectId" viewName="gantt">
		<template #header>
			<card :has-content="false">
				<div class="gantt-options">
					<div class="field">
						<label class="label" for="range">{{ $t('project.gantt.range') }}</label>
						<div class="control">
							<Foo
								ref="flatPickerEl"
								:config="flatPickerConfig"
								class="input"
								id="range"
								:placeholder="$t('project.gantt.range')"
								v-model="flatPickerDateRange"
							/>
						</div>
					</div>
					<div class="field" v-if="!hasDefaultFilters">
						<label class="label" for="range">Reset</label>
						<div class="control">
							<x-button @click="setDefaultFilters">Reset</x-button>
						</div>
					</div>
					<fancycheckbox is-block v-model="filters.showTasksWithoutDates">
						{{ $t('project.gantt.showTasksWithoutDates') }}
					</fancycheckbox>
				</div>
			</card>
		</template>

		<template #default>
			<div class="gantt-chart-container">
				<card :has-content="false" :padding="false" class="has-overflow">
					<gantt-chart
						:filters="filters"
						:tasks="tasks"
						:isLoading="isLoading"
						:default-task-start-date="defaultTaskStartDate"
						:default-task-end-date="defaultTaskEndDate"
						@update:task="updateTask"
					/>
					<TaskForm v-if="canWrite" @create-task="addGanttTask"/>
				</card>
			</div>
		</template>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, ref, toRefs} from 'vue'
import type Flatpickr from 'flatpickr'
import {useI18n} from 'vue-i18n'
import type {RouteLocationNormalized} from 'vue-router'

import {useBaseStore} from '@/stores/base'
import {useAuthStore} from '@/stores/auth'

import Foo from '@/components/misc/flatpickr/Flatpickr.vue'
import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import TaskForm from '@/components/tasks/TaskForm.vue'

import {createAsyncComponent} from '@/helpers/createAsyncComponent'
import {useGanttFilters} from './helpers/useGanttFilters'
import {RIGHTS} from '@/constants/rights'

import type {DateISO} from '@/types/DateISO'
import type {ITask} from '@/modelTypes/ITask'

type Options = Flatpickr.Options.Options

const GanttChart = createAsyncComponent(() => import('@/components/tasks/GanttChart.vue'))

const props = defineProps<{route: RouteLocationNormalized}>()

const baseStore = useBaseStore()
const canWrite = computed(() => baseStore.currentProject?.maxRight > RIGHTS.READ)

const {route} = toRefs(props)
const {
	filters,
	hasDefaultFilters,
	setDefaultFilters,
	tasks,
	isLoading,
	addTask,
	updateTask,
} = useGanttFilters(route)

const DEFAULT_DATE_RANGE_DAYS = 7

const today = new Date()
const defaultTaskStartDate: DateISO = new Date(today.setHours(0, 0, 0, 0)).toISOString()
const defaultTaskEndDate: DateISO = new Date(new Date(
	today.getFullYear(),
	today.getMonth(),
	today.getDate() + DEFAULT_DATE_RANGE_DAYS,
).setHours(23, 59, 0, 0)).toISOString()

async function addGanttTask(title: ITask['title']) {
	return await addTask({
		title,
		projectId: filters.value.projectId,
		startDate: defaultTaskStartDate,
		endDate: defaultTaskEndDate,
	})
}

const flatPickerEl = ref<typeof Foo | null>(null)
const flatPickerDateRange = computed<Date[]>({
	get: () => ([
		new Date(filters.value.dateFrom),
		new Date(filters.value.dateTo),
	]),
	set(newVal) {
		const [dateFrom, dateTo] = newVal.map((date) => date?.toISOString())

		// only set after whole range has been selected
		if (!dateTo) return

		Object.assign(filters.value, {dateFrom, dateTo})
	},
})

const initialDateRange = [filters.value.dateFrom, filters.value.dateTo]

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const flatPickerConfig = computed<Options>(() => ({
	altFormat: t('date.altFormatShort'),
	altInput: true,
	defaultDate: initialDateRange,
	enableTime: false,
	mode: 'range',
	locale: {
		firstDayOfWeek: authStore.settings.weekStart,
	},
}))
</script>

<style lang="scss" scoped>
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
}

:global(.link-share-view:not(.has-background)) .gantt-options {
	border: none;
	box-shadow: none;

	.card-content {
		padding: .5rem;
	}
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

	.select,
	.select select {
		height: auto;
		width: 100%;
		font-size: .8rem;
	}

	.label {
		font-size: .9rem;
	}
}
</style>