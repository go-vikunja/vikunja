<template>
	<ProjectWrapper
		class="project-gantt"
		:is-loading-project="isLoadingProject"
		:project-id="filters.projectId"
		:view-id
	>
		<template #default>
			<Card
				:has-content="false"
				class="gantt-options-card"
			>
				<div class="gantt-options">
					<FormField :label="$t('project.gantt.range')">
						<Foo
							id="range"
							ref="flatPickerEl"
							v-model="flatPickerDateRange"
							:config="flatPickerConfig"
							class="input gantt-range-input"
							:placeholder="$t('project.gantt.range')"
						/>
					</FormField>
					<div
						v-if="!hasDefaultFilters"
						class="field"
					>
						<label
							class="label"
							for="range"
						>Reset</label>
						<div class="control">
							<XButton
								class="gantt-reset-button"
								@click="setDefaultFilters"
							>
								Reset
							</XButton>
						</div>
					</div>
					<FancyCheckbox
						v-model="filters.showTasksWithoutDates"
						is-block
					>
						{{ $t('task.show.noDates') }}
					</FancyCheckbox>
				</div>
			</Card>

			<div class="gantt-chart-container">
				<Card
					:has-content="false"
					:padding="false"
					class="has-overflow gantt-chart-card"
				>
					<GanttChart
						:filters="filters"
						:tasks="tasks"
						:is-loading="isLoading"
						:default-task-start-date="defaultTaskStartDate"
						:default-task-end-date="defaultTaskEndDate"
						@update:task="updateTask"
					/>
					<TaskForm
						v-if="canWrite"
						@createTask="addGanttTask"
					/>
				</Card>
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
import {useFlatpickrLanguage} from '@/helpers/useFlatpickrLanguage'

import Foo from '@/components/misc/flatpickr/Flatpickr.vue'
import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import TaskForm from '@/components/tasks/TaskForm.vue'
import FormField from '@/components/input/FormField.vue'

import GanttChart from '@/components/gantt/GanttChart.vue'
import {useGanttFilters} from '../../../views/project/helpers/useGanttFilters'
import {PERMISSIONS} from '@/constants/permissions'

import type {DateISO} from '@/types/DateISO'
import type {ITask} from '@/modelTypes/ITask'
import type {IProjectView} from '@/modelTypes/IProjectView'

type Options = Flatpickr.Options.Options

const props = defineProps<{
	isLoadingProject: boolean,
	route: RouteLocationNormalized
	viewId: IProjectView['id']
}>()


const baseStore = useBaseStore()
const canWrite = computed(() => baseStore.currentProject?.maxPermission > PERMISSIONS.READ)

const {route, viewId} = toRefs(props)
const {
	filters,
	hasDefaultFilters,
	setDefaultFilters,
	tasks,
	isLoading,
	addTask,
	updateTask,
} = useGanttFilters(route, viewId)

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

const {t} = useI18n({useScope: 'global'})
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatShort'),
	altInput: true,
	defaultDate: [filters.value.dateFrom, filters.value.dateTo],
	enableTime: false,
	mode: 'range',
	locale: useFlatpickrLanguage().value,
} as Options))
</script>

<style lang="scss" scoped>
.gantt-chart-container {
	padding-block-end: 1rem;
	position: relative;
	z-index: 0;
}

.gantt-options-card,
.gantt-chart-card {
	background: var(--vk-bg-panel);
	border: .5px solid var(--vk-border);
	border-radius: 12px;
	box-shadow: none;
}

.gantt-chart-card {
	overflow: hidden;
}

.gantt-options {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-block-end: 1rem;
	padding: .75rem 1rem;
	background: var(--vk-bg);
	border: .5px solid var(--vk-border);
	border-radius: 10px;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

.gantt-range-input {
	background: var(--vk-bg-panel);
	border: .5px solid var(--vk-border-mid);
	border-radius: 10px;
	color: var(--vk-text-secondary);
	font-size: 15px;
	box-shadow: none;

	&::placeholder {
		color: var(--vk-input-placeholder);
	}

	&:focus {
		border-color: var(--vk-accent);
		box-shadow: none;
	}
}

.gantt-reset-button {
	border: .5px dashed var(--vk-action-btn-border) !important;
	background: transparent !important;
	color: var(--vk-text-secondary) !important;
	border-radius: 8px;
	transition: all .12s;
	box-shadow: none !important;

	&:hover {
		border-color: var(--vk-accent) !important;
		background: var(--vk-bg-panel) !important;
		color: var(--vk-text-primary) !important;
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
	margin-block-end: 0;
	inline-size: 33%;

	&:not(:last-child) {
		padding-inline-end: .5rem;
	}

	@media screen and (max-width: $tablet) {
		inline-size: 100%;
		max-inline-size: 100%;
		margin-block-start: .5rem;
		padding-inline-end: 0 !important;
	}

	&, .input {
		font-size: .8rem;
	}

	.select,
	.select select {
		block-size: auto;
		inline-size: 100%;
		font-size: .8rem;
	}

	.label {
		font-size: .9rem;
		color: var(--vk-text-secondary);
	}
}

.project-gantt {
	:deep(.gantt-container),
	:deep(.gantt-chart-wrapper),
	:deep(.gantt-rows-container) {
		background: var(--vk-bg);
	}

	:deep(.fancycheckbox) {
		color: var(--vk-text-secondary);
	}

	:deep(.gantt-timeline) {
		background: var(--vk-bg-panel);
		border-block-end: .5px solid var(--vk-border-mid);
	}

	:deep(.gantt-timeline-months .timeunit-month) {
		background: var(--vk-bg-panel);
		color: var(--vk-text-secondary);
		border-inline-end: .5px solid var(--vk-border-mid);
		font-size: .85rem;
		font-weight: 600;
	}

	:deep(.gantt-timeline-days .timeunit-wrapper) {
		color: #4a4b57;
		font-size: .9rem;

		.weekday {
			color: #3a3b48;
		}

		&.today {
			background: #1e1b3a;
			color: #a78bfa;
			border-radius: 6px 6px 0 0;
		}
	}

	:deep(.gantt-group-band) {
		background: rgb(108 99 245 / 7%);
		border-color: rgb(108 99 245 / 18%);
		border-radius: 8px;
	}

	:deep(.bg-row) {
		background: rgb(19 20 26 / 70%);
	}

	:deep(.bg-row-alt) {
		background: rgb(15 16 22 / 82%);
	}

	:deep(.gantt-vertical-lines line) {
		stroke: var(--vk-border-mid);
		opacity: .75;
	}

	:deep(.gantt-arrow) {
		opacity: .7;
	}

	:deep(.gantt-bar) {
		filter: drop-shadow(0 1px 1px rgb(0 0 0 / 20%));
	}

	:deep(.gantt-bar-text) {
		font-size: 12px;
		font-weight: 500;
		letter-spacing: .01em;
	}

	:deep(.gantt-resize-handle) {
		fill: var(--vk-text-primary);
		stroke: var(--vk-accent);
	}

	:deep(.gantt-collapse-toggle polygon) {
		fill: var(--vk-text-secondary);
	}

	:deep(.task-form .input),
	:deep(.task-form textarea) {
		background: var(--vk-bg-panel);
		border: .5px solid var(--vk-border-mid);
		border-radius: 10px;
		color: var(--vk-text-primary);
		box-shadow: none;

		&::placeholder {
			color: var(--vk-input-placeholder);
		}

		&:focus {
			border-color: #6c63f5;
			box-shadow: none;
		}
	}

	:deep(.task-form .button),
	:deep(.task-form .x-button) {
		background: linear-gradient(135deg, #6c63f5, #8b83f7);
		color: #fff;
		border: 0;
		border-radius: 10px;
		box-shadow: none;

		&:hover {
			opacity: .9;
		}
	}
}
</style>
