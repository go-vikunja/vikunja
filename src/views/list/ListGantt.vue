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
					<gantt-chart
						:filters="filters"
						:tasks="tasks"
						:isLoading="isLoading"
						:default-task-start-date="defaultTaskStartDate"
						:default-task-end-date="defaultTaskEndDate"
						@update:task="updateTask"
					/>
					<TaskForm v-if="canWrite" @create-task="addGanttTask" />
				</card>
			</div>
		</template>
	</ListWrapper>
</template>

<script setup lang="ts">
import {computed, ref, toRefs} from 'vue'
import type Flatpickr from 'flatpickr'
import {useI18n} from 'vue-i18n'
import type {RouteLocationNormalized} from 'vue-router'

import {useBaseStore} from '@/stores/base'
import {useAuthStore} from '@/stores/auth'

import Foo from '@/components/misc/flatpickr/Flatpickr.vue'
import ListWrapper from './ListWrapper.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import TaskForm from '@/components/tasks/TaskForm.vue'

import {createAsyncComponent} from '@/helpers/createAsyncComponent'
import {useGanttFilter} from './helpers/useGanttFilter'
import {RIGHTS} from '@/constants/rights'

import type {DateISO} from '@/types/DateISO'
import type {ITask} from '@/modelTypes/ITask'

type Options = Flatpickr.Options.Options

const GanttChart = createAsyncComponent(() => import('@/components/tasks/gantt-chart.vue'))

const props = defineProps<{route: RouteLocationNormalized}>()

const baseStore = useBaseStore()
const canWrite = computed(() => baseStore.currentList.maxRight > RIGHTS.READ)

const {route} = toRefs(props)
const {
	filters,
	tasks,
	isLoading,
	addTask,
	updateTask,
} = useGanttFilter(route)

const today = new Date(new Date().setHours(0,0,0,0))
const defaultTaskStartDate: DateISO = new Date(today).toISOString()
const defaultTaskEndDate: DateISO = new Date(today.getFullYear(), today.getMonth(), today.getDate() + 7, 23,59,0,0).toISOString()

async function addGanttTask(title: ITask['title']) {
	return await addTask({
		title,
		listId: filters.listId,
		startDate: defaultTaskStartDate,
		endDate: defaultTaskEndDate,
	})
}

const flatPickerEl = ref<typeof Foo | null>(null)
const flatPickerDateRange = computed<Date[]>({
	get: () => ([
		new Date(filters.dateFrom),
		new Date(filters.dateTo),
	]),
	set(newVal) {
		const [dateFrom, dateTo] = newVal.map((date) => date?.toISOString())
		
		// only set after whole range has been selected
		if (!dateTo) return

		Object.assign(filters, {dateFrom, dateTo})
	},
})

const initialDateRange = [filters.dateFrom, filters.dateTo]

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