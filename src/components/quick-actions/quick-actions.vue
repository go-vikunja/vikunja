<template>
	<modal :enabled="active" @close="closeQuickActions" :overflow="isNewTaskCommand">
		<div class="card quick-actions">
			<div class="action-input" :class="{'has-active-cmd': selectedCmd !== null}">
				<div class="active-cmd tag" v-if="selectedCmd !== null">
					{{ selectedCmd.title }}
				</div>
				<input
					v-focus
					class="input"
					:class="{'is-loading': loading}"
					v-model="query"
					:placeholder="placeholder"
					@keyup="search"
					ref="searchInput"
					@keydown.down.prevent="select(0, 0)"
					@keyup.prevent.delete="unselectCmd"
					@keyup.prevent.enter="doCmd"
					@keyup.prevent.esc="closeQuickActions"
				/>
			</div>

			<div class="help has-text-grey-light p-2" v-if="hintText !== '' && !isNewTaskCommand">
				{{ hintText }}
			</div>

			<quick-add-magic class="p-2 modal-container-smaller" v-if="isNewTaskCommand"/>

			<div class="results" v-if="selectedCmd === null">
				<div v-for="(r, k) in results" :key="k" class="result">
					<span class="result-title">
						{{ r.title }}
					</span>
					<div class="result-items">
						<BaseButton
							v-for="(i, key) in r.items"
							:key="key"
							class="result-item-button"
							:class="{'is-strikethrough': (i as DoAction<ITask>)?.done}"
							:ref="(el: Element | ComponentPublicInstance | null) => setResultRefs(el, k, key)"
							@keydown.up.prevent="select(k, key - 1)"
							@keydown.down.prevent="select(k, key + 1)"
							@click.prevent.stop="doAction(r.type, i)"
							@keyup.prevent.enter="doAction(r.type, i)"
							@keyup.prevent.esc="searchInput?.focus()"
						>
							{{ i.title }}
						</BaseButton>
					</div>
				</div>
			</div>
		</div>
	</modal>
</template>

<script setup lang="ts">
import {ref, computed, watchEffect, shallowReactive, type ComponentPublicInstance} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import TaskService from '@/services/task'
import TeamService from '@/services/team'

import TeamModel from '@/models/team'
import ProjectModel from '@/models/project'

import BaseButton from '@/components/base/BaseButton.vue'
import QuickAddMagic from '@/components/tasks/partials/quick-add-magic.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useLabelStore} from '@/stores/labels'
import {useTaskStore} from '@/stores/tasks'

import {getHistory} from '@/modules/projectHistory'
import {parseTaskText, PrefixMode, PREFIXES} from '@/modules/parseTaskText'
import {getQuickAddMagicMode} from '@/helpers/quickAddMagicMode'
import {success} from '@/message'

import type {ITeam} from '@/modelTypes/ITeam'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'

const {t} = useI18n({useScope: 'global'})
const router = useRouter()

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const labelStore = useLabelStore()
const taskStore = useTaskStore()

type DoAction<Type = any> = { type: ACTION_TYPE } & Type

enum ACTION_TYPE {
	CMD = 'cmd',
	TASK = 'task',
	PROJECT = 'project',
	TEAM = 'team',
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
	const { project } = parsedQuery.value
	if (
		searchMode.value === SEARCH_MODE.ALL ||
		searchMode.value === SEARCH_MODE.PROJECTS ||
		project === null
	) {
		return []
	}

	const history = getHistory()
	const allProjects = [
		...new Set([
			...history.map((l) => projectStore.getProjectById(l.id)),
			...projectStore.searchProject(project),
		]),
	]

	return allProjects.filter((l) => typeof l !== 'undefined' && l !== null)
})

// FIXME: use fuzzysearch
const foundCommands = computed(() => availableCmds.value.filter((a) =>
	a.title.toLowerCase().includes(query.value.toLowerCase()),
))

interface Result {
	type: ACTION_TYPE
	title: string
	items: DoAction<any>
}

const results = computed<Result[]>(() => {
	return [
		{
			type: ACTION_TYPE.CMD,
			title: t('quickActions.commands'),
			items: foundCommands.value,
		},
		{
			type: ACTION_TYPE.TASK,
			title: t('quickActions.tasks'),
			items: foundTasks.value,
		},
		{
			type: ACTION_TYPE.PROJECT,
			title: t('quickActions.projects'),
			items: foundProjects.value,
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

const currentProject = computed(() => Object.keys(baseStore.currentProject).length === 0
	? null
	: baseStore.currentProject,
)

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
		PREFIXES[getQuickAddMagicMode()] ?? PREFIXES[PrefixMode.Default]
	return t('quickActions.hint', prefixes)
})

const availableCmds = computed(() => {
	const cmds = []
	if (currentProject.value !== null) {
		cmds.push(commands.value.newTask, commands.value.newProject)
	}
	cmds.push(commands.value.newTeam)
	return cmds
})

const parsedQuery = computed(() => parseTaskText(query.value, getQuickAddMagicMode()))

const searchMode = computed(() => {
	if (query.value === '') {
		return SEARCH_MODE.ALL
	}
	const { text, project, labels, assignees } = parsedQuery.value
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

type Filter = {by: string, value: string | number, comparator: string}

function filtersToParams(filters: Filter[]) {
	const filter_by : Filter['by'][] = []
	const filter_value : Filter['value'][] = []
	const filter_comparator : Filter['comparator'][] = []

	filters.forEach(({by, value, comparator}) => {
		filter_by.push(by)
		filter_value.push(value)
		filter_comparator.push(comparator)
	})

	return {
		filter_by,
		filter_value,
		filter_comparator,
	}
}

function searchTasks() {
	if (
		searchMode.value !== SEARCH_MODE.ALL &&
		searchMode.value !== SEARCH_MODE.TASKS
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

	const { text, project: projectName, labels } = parsedQuery.value

	const filters: Filter[] = []

	// FIXME: improve types
	function addFilter(
		by: Filter['by'],
		value: Filter['value'],
		comparator: Filter['comparator'],
	) {
		filters.push({
			by,
			value,
			comparator,
		})
	}

	if (projectName !== null) {
		const project = projectStore.findProjectByExactname(projectName)
		if (project !== null) {
			addFilter('projectId', project.id, 'equals')
		}
	}

	if (labels.length > 0) {
		const labelIds = labelStore.getLabelsByExactTitles(labels).map((l) => l.id)
		if (labelIds.length > 0) {
			addFilter('labels', labelIds.join(), 'in')
		}
	}

		const params = {
			s: text,
			...filtersToParams(filters),
		}

	taskSearchTimeout.value = setTimeout(async () => {
		const r = await taskService.getAll({}, params) as  DoAction<ITask>[]
		foundTasks.value = r.map((t) => {
			t.type = ACTION_TYPE.TASK
			const project = projectStore.getProjectById(t.projectId)
			if (project !== null) {
				t.title = `${t.title} (${project.title})`
			}
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
	const { assignees } = parsedQuery.value
	teamSearchTimeout.value = setTimeout(async () => {
		const teamSearchPromises = assignees.map((t) =>
			teamService.getAll({}, { s: t }),
		)
		const teamsResult = await Promise.all(teamSearchPromises)
		foundTeams.value = teamsResult.flat().map((team) => {
			team.title = team.name
			return team
		})
	}, 150)
}

function search() {
	searchTasks()
	searchTeams()
}

const searchInput = ref<HTMLElement | null>(null)

async function doAction(type: ACTION_TYPE, item: DoAction) {
	switch (type) {
		case ACTION_TYPE.PROJECT:
			closeQuickActions()
			await router.push({
				name: 'project.index',
				params: { projectId: (item as DoAction<IProject>).id },
			})
			break
		case ACTION_TYPE.TASK:
			closeQuickActions()
			await router.push({
				name: 'task.detail',
				params: { id: (item as DoAction<ITask>).id },
			})
			break
		case ACTION_TYPE.TEAM:
			closeQuickActions()
			await router.push({
				name: 'teams.edit',
				params: { id: (item as DoAction<ITeam>).id },
			})
			break
		case ACTION_TYPE.CMD:
			query.value = ''
			selectedCmd.value = item as DoAction<Command>
			searchInput.value?.focus()
			break
	}
}

async function doCmd() {
	if (results.value.length === 1 && results.value[0].items.length === 1) {
		const result = results.value[0]
		doAction(result.type, result.items[0])
		return
	}

	if (selectedCmd.value === null || query.value === '') {
		return
	}

	closeQuickActions()
	await selectedCmd.value.action()
}

async function newTask() {
	if (currentProject.value === null) {
		return
	}
	const task = await taskStore.createNewTask({
		title: query.value,
		projectId: currentProject.value.id,
	})
	success({ message: t('task.createSuccess') })
	await router.push({ name: 'task.detail', params: { id: task.id } })
}

async function newProject() {
	if (currentProject.value === null) {
		return
	}
	const newProject = await projectStore.createProject(new ProjectModel({
		title: query.value,
	}))
	success({ message: t('project.create.createdSuccess')})
	await router.push({
		name: 'project.index',
		params: { projectId: newProject.id },
	})
}

async function newTeam() {
	const newTeam = new TeamModel({ name: query.value })
	const team = await teamService.create(newTeam)
	await router.push({
		name: 'teams.edit',
		params: { id: team.id },
	})
	success({ message: t('team.create.success') })
}

type BaseButtonInstance = InstanceType<typeof BaseButton>
const resultRefs = ref<(BaseButtonInstance | null)[][]>([])

function setResultRefs(el: Element | ComponentPublicInstance | null, index: number, key: number) {
	if (resultRefs.value[index] === undefined) {
		resultRefs.value[index] = []
	}

	resultRefs.value[index][key] =  el as (BaseButtonInstance | null)
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
	let elems = resultRefs.value[parentIndex][index]
	if (results.value[parentIndex].items.length === index) {
		elems = resultRefs.value[parentIndex + 1] ? resultRefs.value[parentIndex + 1][0] : undefined
	}
	if (
		typeof elems === 'undefined'
		/* || elems.length === 0 */
	) {
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
</script>

<style lang="scss" scoped>
.quick-actions {
	overflow: hidden;
	
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
	}

	&.has-active-cmd .input {
		padding-left: .5rem;
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
}

// HACK:
// FIXME:
.modal-container-smaller :deep(.hint-modal .modal-container) {
	height: calc(100vh - 5rem);
}
</style>