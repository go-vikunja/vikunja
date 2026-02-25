<template>
	<ProjectWrapper
		class="project-gantt"
		:is-loading-project="isLoadingProject"
		:project-id="filters.projectId"
		:view-id
	>
		<template #default>
			<Card :has-content="false">
				<div class="gantt-options">
					<FormField :label="$t('project.gantt.range')">
						<Foo
							id="range"
							ref="flatPickerEl"
							v-model="flatPickerDateRange"
							:config="flatPickerConfig"
							class="input"
							:placeholder="$t('project.gantt.range')"
						/>
					</FormField>
					<FancyCheckbox
						v-model="filters.showTasksWithoutDates"
						is-block
					>
						{{ $t('task.show.noDates') }}
					</FancyCheckbox>
					<FancyCheckbox
						v-model="filters.showDoneTasks"
						is-block
					>
						{{ $t('task.show.completed') }}
					</FancyCheckbox>
					<SubprojectFilter
						:project-id="filters.projectId"
						:show-legend="true"
						@update:includeSubprojects="onSubprojectToggle"
						@update:excludeProjectIds="onExcludeChange"
						@update:colorMap="onColorMapChange"
					/>
					<GanttArrowSettings />
				</div>
			</Card>

			<div class="gantt-chart-container">
				<Card
					:has-content="false"
					:padding="false"
					class="has-overflow"
				>
					<GanttChart
						:filters="filters"
						:tasks="tasks"
						:is-loading="isLoading"
						:default-task-start-date="defaultTaskStartDate"
						:default-task-end-date="defaultTaskEndDate"
						:subproject-color-map="subprojectColorMap"
						@update:task="updateTask"
					/>
					<div class="gantt-bottom-bar">
						<TaskForm
							v-if="canWrite"
							@createTask="addGanttTask"
						/>
						<XButton
							v-if="canWrite"
							variant="primary"
							icon="layer-group"
							class="gantt-action-btn"
							@click="showCreateFromTemplateModal = true"
						>
							{{ $t('task.template.fromTemplate') }}
						</XButton>
						<XButton
							v-if="canWrite"
							variant="primary"
							icon="link"
							class="gantt-action-btn"
							@click="showCreateFromChainModal = true"
						>
							{{ $t('task.chain.createFromChain') }}
						</XButton>
					</div>
				</Card>
			</div>
		</template>
	</ProjectWrapper>

	<CreateFromTemplateModal
		:enabled="showCreateFromTemplateModal"
		:default-project-id="filters.projectId"
		@close="showCreateFromTemplateModal = false"
		@created="onTaskCreatedFromTemplate"
	/>
	<CreateFromChainModal
		:enabled="showCreateFromChainModal"
		:project-id="filters.projectId"
		@close="showCreateFromChainModal = false"
		@created="loadTasks()"
	/>
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
import SubprojectFilter from '@/components/project/partials/SubprojectFilter.vue'
import GanttArrowSettings from '@/components/gantt/GanttArrowSettings.vue'
import CreateFromTemplateModal from '@/components/tasks/partials/CreateFromTemplateModal.vue'
import CreateFromChainModal from '@/components/tasks/partials/CreateFromChainModal.vue'
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

const subprojectParams = ref<Record<string, unknown>>({})
const subprojectColorMap = ref<Map<number, string>>(new Map())
const showCreateFromTemplateModal = ref(false)
const showCreateFromChainModal = ref(false)

function onSubprojectToggle(enabled: boolean) {
	if (enabled) {
		subprojectParams.value = {...subprojectParams.value, include_subprojects: true}
	} else {
		const {include_subprojects, exclude_project_ids, ...rest} = subprojectParams.value
		subprojectParams.value = rest
	}
	loadTasks()
}

function onExcludeChange(ids: string) {
	if (ids) {
		subprojectParams.value = {...subprojectParams.value, exclude_project_ids: ids}
	} else {
		const {exclude_project_ids, ...rest} = subprojectParams.value
		subprojectParams.value = rest
	}
	loadTasks()
}

function onColorMapChange(map: Map<number, string>) {
	subprojectColorMap.value = map
}

function onTaskCreatedFromTemplate(createdTask: ITask) {
	if (createdTask.projectId === filters.value.projectId) {
		loadTasks()
	}
}

const {
	filters,
	tasks,
	isLoading,
	addTask,
	updateTask,
	loadTasks,
} = useGanttFilters(route, viewId, subprojectParams)

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
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatShort'),
	altInput: true,
	defaultDate: initialDateRange,
	enableTime: false,
	mode: 'range',
	locale: useFlatpickrLanguage().value,
} as Options))
</script>

<style lang="scss" scoped>
.gantt-chart-container {
	padding-block-end: 1rem;
}

.gantt-bottom-bar {
	display: flex;
	align-items: center;
	gap: .5rem;
	padding: .5rem;
	flex-wrap: wrap;

	:deep(.add-new-task) {
		padding: 0;
		margin: 0;

		.button {
			font-size: .8rem;
			padding-block: .4rem;
			padding-inline: .75rem;
		}
	}
}

.gantt-action-btn {
	font-size: .8rem;
	padding-block: .4rem;
	padding-inline: .75rem;
}

.gantt-options {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-block-end: 1rem;

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
	}
}
</style>
