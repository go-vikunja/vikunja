<template>
	<Modal
		:enabled="active"
		:overflow="isNewTaskCommand"
		@close="closeQuickActions"
	>
		<div class="card quick-actions">
			<div
				class="action-input"
				:class="{'has-active-cmd': selectedCmd !== null}"
			>
				<div
					v-if="selectedCmd !== null"
					class="active-cmd tag"
				>
					{{ selectedCmd.title }}
				</div>
				<input
					ref="searchInput"
					v-model="query"
					v-focus
					class="input"
					:class="{'is-loading': loading}"
					:placeholder="placeholder"
					@keyup="search"
					@keydown.down.prevent="select(0, 0)"
					@keyup.prevent.delete="unselectCmd"
					@keyup.prevent.enter="doCmd"
					@keyup.prevent.esc="closeQuickActions"
				>
				<BaseButton
					class="close"
					@click="closeQuickActions"
				>
					<Icon icon="times" />
				</BaseButton>
			</div>

			<div
				v-if="hintText !== '' && !isNewTaskCommand"
				class="help has-text-grey-light p-2"
			>
				{{ hintText }}
			</div>

			<QuickAddMagic v-if="isNewTaskCommand" />

			<div
				v-if="selectedCmd === null"
				class="results"
			>
				<div
					v-for="(r, k) in results"
					:key="k"
					class="result"
				>
					<span class="result-title">
						{{ r.title }}
					</span>
					<div class="result-items">
						<BaseButton
							v-for="(i, key) in r.items"
							:key="key"
							:ref="(el: Element | ComponentPublicInstance | null) => setResultRefs(el, k, Number(key))"
							class="result-item-button"
							:class="{'is-strikethrough': isDoneItem(i as SearchResult | Command)}"
							@keydown.up.prevent="select(k, Number(key) - 1)"
							@keydown.down.prevent="select(k, Number(key) + 1)"
							@click.prevent.stop="doAction(r.type, i as SearchResult | Command)"
							@keyup.prevent.enter="doAction(r.type, i as SearchResult | Command)"
							@keyup.prevent.esc="searchInput?.focus()"
						>
							<template v-if="r.type === ACTION_TYPE.LABELS && i">
								<XLabel :label="i as unknown as ILabel" />
							</template>
							<template v-else-if="r.type === ACTION_TYPE.TASK && i">
								<SingleTaskInlineReadonly
									:task="i as unknown as ITask"
									:show-project="true"
								/>
							</template>
							<template v-else>
								<span
									v-if="isSavedFilterItem(i as SearchResult | Command)"
									class="saved-filter-icon icon"
								>
									<Icon icon="filter" />
								</span>
								{{ getItemTitle(i as SearchResult | Command) }}
							</template>
						</BaseButton>
					</div>
				</div>
			</div>
		</div>
	</Modal>
</template>

<script setup lang="ts">
import {type ComponentPublicInstance, computed, ref, shallowReactive, watchEffect} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import TaskService from '@/services/task'
import TeamService from '@/services/team'

import TeamModel from '@/models/team'
import ProjectModel from '@/models/project'
import TaskModel from '@/models/task'

import BaseButton from '@/components/base/BaseButton.vue'
import QuickAddMagic from '@/components/tasks/partials/QuickAddMagic.vue'
import XLabel from '@/components/tasks/partials/Label.vue'
import SingleTaskInlineReadonly from '@/components/tasks/partials/SingleTaskInlineReadonly.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useLabelStore} from '@/stores/labels'
import {useTaskStore} from '@/stores/tasks'
import {useAuthStore} from '@/stores/auth'

import {getHistory} from '@/modules/projectHistory'
import {parseTaskText, PREFIXES, PrefixMode} from '@/modules/parseTaskText'
import {success} from '@/message'

import type {ITeam} from '@/modelTypes/ITeam'
import type {ITask} from '@/modelTypes/ITask'
import type {ILabel} from '@/modelTypes/ILabel'
import {isSavedFilter} from '@/services/savedFilter'

const {t} = useI18n({useScope: 'global'})
const router = useRouter()

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const labelStore = useLabelStore()
const taskStore = useTaskStore()
const authStore = useAuthStore()

type DoAction<Type> = { type: ACTION_TYPE } & Type

interface SearchResult {
	id: number
	title: string
	done?: boolean
	[key: string]: unknown
}

enum ACTION_TYPE {
	CMD = 'cmd',
	TASK = 'task',
	PROJECT = 'project',
	TEAM = 'team',
	LABELS = 'labels',
}

enum COMMAND_TYPE {
	NEW_TASK = 'newTask',
	NEW_PROJECT = 'newProject',
	NEW_TEAM = 'newTeam',
}

enum SEARCH_MODE {
	ALL = 'all',
	TASKS = 'tasks',
	PROJECTS = 'projects',
	TEAMS = 'teams',
}

const query = ref('')
const selectedCmd = ref<Command | null>(null)

const foundTasks = ref<DoAction<ITask>[]>([])
const taskService = shallowReactive(new TaskService())

const foundTeams = ref<ITeam[]>([])
const teamService = shallowReactive(new TeamService())

const active = computed(() => baseStore.quickActionsActive)

watchEffect(() => {
	if (!active.value) {
		reset()
	}
})

function closeQuickActions() {
	baseStore.setQuickActionsActive(false)
}

const foundProjects = computed(() => {
	const {project, text, labels, assignees} = parsedQuery.value

	if (project !== null) {
		return projectStore.searchProjectAndFilter(project ?? text)
			.filter(p => Boolean(p))
	}

	if (labels.length > 0 || assignees.length > 0) {
		return []
	}

	if (text === '') {
		const history = getHistory()
		return history.map((p) => projectStore.projects[p.id])
			.filter(p => Boolean(p))
	}

	return projectStore.searchProjectAndFilter(project ?? text)
		.filter(p => Boolean(p))
})

const foundLabels = computed(() => {
	const {labels, text} = parsedQuery.value
	if (text === '' && labels.length === 0) {
		return []
	}

	if (labels.length > 0) {
		return labelStore.filterLabelsByQuery([], labels[0])
	}

	return labelStore.filterLabelsByQuery([], text)
})

// FIXME: use fuzzysearch
const foundCommands = computed(() => availableCmds.value.filter((a) =>
	a.title.toLowerCase().includes(query.value.toLowerCase()),
))

interface Result {
	type: ACTION_TYPE
	title: string
	items: SearchResult[] | Command[]
}

const results = computed(() => {
	return [
		{
			type: ACTION_TYPE.CMD,
			title: t('quickActions.commands'),
			items: foundCommands.value,
		},
		{
			type: ACTION_TYPE.PROJECT,
			title: t('quickActions.projects'),
			items: foundProjects.value,
		},
		{
			type: ACTION_TYPE.TASK,
			title: t('quickActions.tasks'),
			items: foundTasks.value,
		},
		{
			type: ACTION_TYPE.LABELS,
			title: t('quickActions.labels'),
			items: foundLabels.value,
		},
		{
			type: ACTION_TYPE.TEAM,
			title: t('quickActions.teams'),
			items: foundTeams.value,
		},
	].filter((i) => i.items.length > 0)
})

const loading = computed(() =>
	taskService.loading ||
	projectStore.isLoading ||
	teamService.loading,
)

interface Command {
	type: COMMAND_TYPE
	title: string
	placeholder: string
	action: () => Promise<void>
}

const commands = computed<{ [key in COMMAND_TYPE]: Command }>(() => ({
	newTask: {
		type: COMMAND_TYPE.NEW_TASK,
		title: t('quickActions.cmds.newTask'),
		placeholder: t('quickActions.newTask'),
		action: newTask,
	},
	newProject: {
		type: COMMAND_TYPE.NEW_PROJECT,
		title: t('quickActions.cmds.newProject'),
		placeholder: t('quickActions.newProject'),
		action: newProject,
	},
	newTeam: {
		type: COMMAND_TYPE.NEW_TEAM,
		title: t('quickActions.cmds.newTeam'),
		placeholder: t('quickActions.newTeam'),
		action: newTeam,
	},
}))

const placeholder = computed(() => selectedCmd.value?.placeholder || t('quickActions.placeholder'))

const currentProject = computed(() => {
	if (!baseStore.currentProject || Object.keys(baseStore.currentProject).length === 0 || isSavedFilter(baseStore.currentProject)) {
		return null
	}
	
	return baseStore.currentProject
})

const hintText = computed(() => {
	if (selectedCmd.value !== null && currentProject.value !== null) {
		switch (selectedCmd.value.type) {
			case COMMAND_TYPE.NEW_TASK:
				return t('quickActions.createTask', {
					title: currentProject.value.title,
				})
			case COMMAND_TYPE.NEW_PROJECT:
				return t('quickActions.createProject')
		}
	}
	const prefixes =
		PREFIXES[authStore.settings.frontendSettings.quickAddMagicMode] ?? PREFIXES[PrefixMode.Default]
	return prefixes ? t('quickActions.hint', prefixes as unknown as Record<string, unknown>) : t('quickActions.hint')
})

const availableCmds = computed(() => {
	return [
		commands.value.newTask,
		commands.value.newProject,
		commands.value.newTeam,
	]
})

const parsedQuery = computed(() => parseTaskText(query.value, authStore.settings.frontendSettings.quickAddMagicMode))

const searchMode = computed(() => {
	if (query.value === '') {
		return SEARCH_MODE.ALL
	}

	const {text, project, labels, assignees} = parsedQuery.value
	if (assignees.length === 0 && text !== '') {
		return SEARCH_MODE.TASKS
	}

	if (
		assignees.length === 0 &&
		project !== null &&
		text === '' &&
		labels.length === 0
	) {
		return SEARCH_MODE.PROJECTS
	}

	if (
		assignees.length > 0 &&
		project === null &&
		text === '' &&
		labels.length === 0
	) {
		return SEARCH_MODE.TEAMS
	}

	return SEARCH_MODE.ALL
})

const isNewTaskCommand = computed(() => (
	selectedCmd.value !== null &&
	selectedCmd.value.type === COMMAND_TYPE.NEW_TASK
))

const taskSearchTimeout = ref<ReturnType<typeof setTimeout> | null>(null)

function searchTasks() {
	if (
		searchMode.value !== SEARCH_MODE.ALL &&
		searchMode.value !== SEARCH_MODE.TASKS &&
		searchMode.value !== SEARCH_MODE.PROJECTS
	) {
		foundTasks.value = []
		return
	}

	if (selectedCmd.value !== null) {
		return
	}

	if (taskSearchTimeout.value !== null) {
		clearTimeout(taskSearchTimeout.value)
		taskSearchTimeout.value = null
	}

	const {text, project: projectName, labels} = parsedQuery.value

	let filter = ''

	if (projectName !== null) {
		const project = projectStore.findProjectByExactname(projectName)
		console.log({project})
		if (project !== null) {
			filter += ' project = ' + project.id
		}
	}

	if (labels.length > 0) {
		const labelIds = labelStore.getLabelsByExactTitles(labels).map((l) => l.id)
		if (labelIds.length > 0) {
			filter += 'labels in ' + labelIds.join(', ')
		}
	}

	const params = {
		s: text,
		sort_by: 'done',
		filter,
	}

	taskSearchTimeout.value = setTimeout(async () => {
		const r = await taskService.getAll(new TaskModel(), params) as DoAction<ITask>[]
		foundTasks.value = r.map((t) => {
			t.type = ACTION_TYPE.TASK
			return t
		})
	}, 150)
}

const teamSearchTimeout = ref<ReturnType<typeof setTimeout> | null>(null)

function searchTeams() {
	if (
		searchMode.value !== SEARCH_MODE.ALL &&
		searchMode.value !== SEARCH_MODE.TEAMS
	) {
		foundTeams.value = []
		return
	}
	if (query.value === '' || selectedCmd.value !== null) {
		return
	}
	if (teamSearchTimeout.value !== null) {
		clearTimeout(teamSearchTimeout.value)
		teamSearchTimeout.value = null
	}
	const {assignees} = parsedQuery.value
	teamSearchTimeout.value = setTimeout(async () => {
		const teamSearchPromises = assignees.map((t) =>
			teamService.getAll(new TeamModel(), {s: t}),
		)
		const teamsResult = await Promise.all(teamSearchPromises)
		foundTeams.value = teamsResult.flat().map((team) => {
			return {
				...team,
				title: team.name,
			}
		})
	}, 150)
}

function search() {
	searchTasks()
	searchTeams()
}

const searchInput = ref<HTMLElement | null>(null)

async function doAction(type: ACTION_TYPE, item: SearchResult | Command) {
	// Handle commands differently
	if ('action' in item && typeof item.action === 'function') {
		closeQuickActions()
		await item.action()
		return
	}
	
	// If we get here, item is a SearchResult
	const searchResult = item as SearchResult
	switch (type) {
		case ACTION_TYPE.PROJECT:
			closeQuickActions()
			await router.push({
				name: 'project.index',
				params: {projectId: searchResult.id},
			})
			break
		case ACTION_TYPE.TASK:
			closeQuickActions()
			await router.push({
				name: 'task.detail',
				params: {id: searchResult.id},
			})
			break
		case ACTION_TYPE.TEAM:
			closeQuickActions()
			await router.push({
				name: 'teams.edit',
				params: {id: searchResult.id},
			})
			break
		case ACTION_TYPE.CMD:
			query.value = ''
			selectedCmd.value = item as unknown as Command
			searchInput.value?.focus()
			break
		case ACTION_TYPE.LABELS: {
			const label = item as {title: string}
			if (/\s/.test(label.title)) {
				query.value = '*"' + label.title + '"'
			} else {
				query.value = '*' + label.title
			}
			searchInput.value?.focus()
			searchTasks()
			break
		}
	}
}

async function doCmd() {
	if (results.value.length === 1 && results.value[0].items.length === 1) {
		const result = results.value[0]
		doAction(result.type, result.items[0] as SearchResult)
		return
	}

	if (selectedCmd.value === null || query.value === '') {
		return
	}

	closeQuickActions()
	await selectedCmd.value.action()
}

async function newTask() {
	let projectId = authStore.settings.defaultProjectId
	if (currentProject.value?.id && currentProject.value.id > 0) {
		projectId = currentProject.value.id
	}
	const task = await taskStore.createNewTask({
		title: query.value,
		projectId,
	})
	success({message: t('task.createSuccess')})
	await router.push({name: 'task.detail', params: {id: task.id}})
}

async function newProject() {
	const parentProjectId = currentProject.value?.id ?? 0
	await projectStore.createProject(new ProjectModel({
		title: query.value,
		parentProjectId: Math.max(parentProjectId, 0),
	}))
	success({message: t('project.create.createdSuccess')})
}

async function newTeam() {
	const newTeam = new TeamModel({name: query.value})
	const team = await teamService.create(newTeam)
	await router.push({
		name: 'teams.edit',
		params: {id: team.id},
	})
	success({message: t('team.create.success')})
}

type BaseButtonInstance = InstanceType<typeof BaseButton>
const resultRefs = ref<(BaseButtonInstance | null)[][]>([])

function setResultRefs(el: Element | ComponentPublicInstance | null, index: number, key: number) {
	if (resultRefs.value[index] === undefined) {
		resultRefs.value[index] = []
	}

	resultRefs.value[index][key] = el as (BaseButtonInstance | null)
}

function select(parentIndex: number, index: number) {
	if (index < 0 && parentIndex === 0) {
		searchInput.value?.focus()
		return
	}
	if (index < 0) {
		parentIndex--
		index = results.value[parentIndex].items.length - 1
	}
	let elems: BaseButtonInstance | null = resultRefs.value[parentIndex]?.[index] ?? null
	if (results.value[parentIndex].items.length === index) {
		elems = resultRefs.value[parentIndex + 1]?.[0] ?? null
	}
	if (elems === null) {
		return
	}
	if (Array.isArray(elems)) {
		elems[0].focus()
		return
	}
	elems?.focus()
}

function unselectCmd() {
	if (query.value !== '') {
		return
	}
	selectedCmd.value = null
}

function reset() {
	query.value = ''
	selectedCmd.value = null
}

// Helper functions for template
function isDoneItem(item: SearchResult | Command): boolean {
	if ('action' in item) {
		// This is a Command, commands are never "done"
		return false
	}
	return item && typeof item === 'object' && 'done' in item && Boolean(item.done)
}

function isSavedFilterItem(item: SearchResult | Command): boolean {
	if ('action' in item) {
		// Commands are not saved filter items
		return false
	}
	return item && typeof item === 'object' && 'id' in item && typeof item.id === 'number' && item.id < -1
}

function getItemTitle(item: SearchResult | Command): string {
	return item && typeof item === 'object' && 'title' in item && typeof item.title === 'string' ? item.title : ''
}
</script>

<style lang="scss" scoped>
.quick-actions {
	overflow: hidden;
	justify-content: flex-start !important;

	// FIXME: changed position should be an option of the modal
	:deep(.modal-content) {
		top: 3rem;
		transform: translate(-50%, 0);
	}
}

.action-input {
	display: flex;
	align-items: center;

	.input {
		border: 0;
		font-size: 1.5rem;
		
		@media screen and (max-width: $tablet) {
			padding-right: .25rem;
		}
	}

	&.has-active-cmd .input {
		padding-left: .5rem;
	}

	.close {
		padding: 0 1rem 0 .5rem;
		font-size: 1.5rem;
		
		@media screen and (min-width: $tablet + 1) {
			display: none;
		}
	}
}

.active-cmd {
	font-size: 1.25rem;
	margin-left: .5rem;
	background-color: var(--grey-100);
	color: var(--grey-800);
}

.results {
	text-align: left;
	width: 100%;
	color: var(--grey-800);
}

.result-title {
	background: var(--grey-100);
	padding: .5rem;
	display: block;
	font-size: .75rem;
}

.result-item-button {
	font-size: .9rem;
	width: 100%;
	background: transparent;
	color: var(--grey-800);
	text-align: left;
	box-shadow: none;
	border-radius: 0;
	text-transform: none;
	font-family: $family-sans-serif;
	font-weight: normal;
	padding: .5rem .75rem;
	border: none;
	cursor: pointer;

	&:focus,
	&:hover {
		background: var(--grey-100);
		box-shadow: none !important;
	}

	&:active {
		background: var(--grey-100);
	}
	
	.saved-filter-icon {
		font-size: .75rem;
		width: .75rem;
		margin-right: .25rem;
		color: var(--grey-400)
	}
	
	&:has(.saved-filter-icon) {
		display: inline-flex;
		align-items: center;
	}
}
</style>
