<template>
	<modal v-if="active" class="quick-actions" @close="closeQuickActions" :overflow="isNewTaskCommand">
		<div class="card">
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
					@keydown.down.prevent="() => select(0, 0)"
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
						<button
							v-for="(i, key) in r.items"
							:key="key"
							:ref="`result-${k}_${key}`"
							@keydown.up.prevent="() => select(k, key - 1)"
							@keydown.down.prevent="() => select(k, key + 1)"
							@click.prevent.stop="() => doAction(r.type, i)"
							@keyup.prevent.enter="() => doAction(r.type, i)"
							@keyup.prevent.esc="() => $refs.searchInput.focus()"
							:class="{'is-strikethrough': i.done}"
						>
							{{ i.title }}
						</button>
					</div>
				</div>
			</div>
		</div>
	</modal>
</template>

<script>
import TaskService from '@/services/task'
import TeamService from '@/services/team'

import NamespaceModel from '@/models/namespace'
import TeamModel from '@/models/team'

import {CURRENT_LIST, LOADING, LOADING_MODULE, QUICK_ACTIONS_ACTIVE} from '@/store/mutation-types'
import ListModel from '@/models/list'
import QuickAddMagic from '@/components/tasks/partials/quick-add-magic.vue'
import {getHistory} from '@/modules/listHistory'
import {parseTaskText, PrefixMode} from '@/modules/parseTaskText'
import {getQuickAddMagicMode} from '@/helpers/quickAddMagicMode'
import {PREFIXES} from '@/modules/parseTaskText'

const TYPE_LIST = 'list'
const TYPE_TASK = 'task'
const TYPE_CMD = 'cmd'
const TYPE_TEAM = 'team'

const CMD_NEW_TASK = 'newTask'
const CMD_NEW_LIST = 'newList'
const CMD_NEW_NAMESPACE = 'newNamespace'
const CMD_NEW_TEAM = 'newTeam'

const SEARCH_MODE_ALL = 'all'
const SEARCH_MODE_TASKS = 'tasks'
const SEARCH_MODE_LISTS = 'lists'
const SEARCH_MODE_TEAMS = 'teams'

export default {
	name: 'quick-actions',
	components: {QuickAddMagic},
	data() {
		return {
			query: '',
			selectedCmd: null,

			foundTasks: [],
			taskSearchTimeout: null,
			taskService: new TaskService(),

			foundTeams: [],
			teamSearchTimeout: null,
			teamService: new TeamService(),
		}
	},
	computed: {
		active() {
			const active = this.$store.state[QUICK_ACTIONS_ACTIVE]
			if (!active) {
				this.reset()
			}
			return active
		},
		results() {
			let lists = []
			if (this.searchMode === SEARCH_MODE_ALL || this.searchMode === SEARCH_MODE_LISTS) {
				const {list} = this.parsedQuery

				if (list === null) {
					lists = []
				} else {
					const ncache = {}
					const history = getHistory()
					// Puts recently visited lists at the top
					const allLists = [...new Set([
						...history.map(l => {
							return this.$store.getters['lists/getListById'](l.id)
						}),
						...this.$store.getters['lists/searchList'](list),
					])]

					lists = allLists.filter(l => {
						if (typeof l === 'undefined' || l === null) {
							return false
						}

						if (typeof ncache[l.namespaceId] === 'undefined') {
							ncache[l.namespaceId] = this.$store.getters['namespaces/getNamespaceById'](l.namespaceId)
						}

						return !ncache[l.namespaceId].isArchived
					})
				}
			}

			const cmds = this.availableCmds
				.filter(a => a.title.toLowerCase().includes(this.query.toLowerCase()))

			return [
				{
					type: TYPE_CMD,
					title: this.$t('quickActions.commands'),
					items: cmds,
				},
				{
					type: TYPE_TASK,
					title: this.$t('quickActions.tasks'),
					items: this.foundTasks,
				},
				{
					type: TYPE_LIST,
					title: this.$t('quickActions.lists'),
					items: lists,
				},
				{
					type: TYPE_TEAM,
					title: this.$t('quickActions.teams'),
					items: this.foundTeams,
				},
			].filter(i => i.items.length > 0)
		},
		nothing() {
			return this.search === '' || Object.keys(this.results).length === 0
		},
		loading() {
			return this.taskService.loading ||
				(this.$store.state[LOADING] && this.$store.state[LOADING_MODULE] === 'namespaces') ||
				(this.$store.state[LOADING] && this.$store.state[LOADING_MODULE] === 'lists') ||
				this.teamService.loading
		},
		placeholder() {
			if (this.selectedCmd !== null) {
				switch (this.selectedCmd.action) {
					case CMD_NEW_TASK:
						return this.$t('quickActions.newTask')
					case CMD_NEW_LIST:
						return this.$t('quickActions.newList')
					case CMD_NEW_NAMESPACE:
						return this.$t('quickActions.newNamespace')
					case CMD_NEW_TEAM:
						return this.$t('quickActions.newTeam')
				}
			}

			return this.$t('quickActions.placeholder')
		},
		hintText() {
			let namespace

			if (this.selectedCmd !== null && this.currentList !== null) {
				switch (this.selectedCmd.action) {
					case CMD_NEW_TASK:
						return this.$t('quickActions.createTask', {title: this.currentList.title})
					case CMD_NEW_LIST:
						namespace = this.$store.getters['namespaces/getNamespaceById'](this.currentList.namespaceId)
						return this.$t('quickActions.createList', {title: namespace.title})
				}
			}

			const prefixes = PREFIXES[getQuickAddMagicMode()] ?? PREFIXES[PrefixMode.Default]

			return this.$t('quickActions.hint', prefixes)
		},
		currentList() {
			return Object.keys(this.$store.state[CURRENT_LIST]).length === 0 ? null : this.$store.state[CURRENT_LIST]
		},
		availableCmds() {
			const cmds = []

			if (this.currentList !== null) {
				cmds.push({
					title: this.$t('quickActions.cmds.newTask'),
					action: CMD_NEW_TASK,
				})
				cmds.push({
					title: this.$t('quickActions.cmds.newList'),
					action: CMD_NEW_LIST,
				})
			}
			cmds.push({
				title: this.$t('quickActions.cmds.newNamespace'),
				action: CMD_NEW_NAMESPACE,
			})
			cmds.push({
				title: this.$t('quickActions.cmds.newTeam'),
				action: CMD_NEW_TEAM,
			})

			return cmds
		},
		parsedQuery() {
			return parseTaskText(this.query, getQuickAddMagicMode())
		},
		searchMode() {
			if (this.query === '') {
				return SEARCH_MODE_ALL
			}

			const {text, list, labels, assignees} = this.parsedQuery

			if (assignees.length === 0 && text !== '') {
				return SEARCH_MODE_TASKS
			}
			if (assignees.length === 0 && list !== null && text === '' && labels.length === 0) {
				return SEARCH_MODE_LISTS
			}
			if (assignees.length > 0 && list === null && text === '' && labels.length === 0) {
				return SEARCH_MODE_TEAMS
			}

			return SEARCH_MODE_ALL
		},
		isNewTaskCommand() {
			return this.selectedCmd !== null && this.selectedCmd.action === CMD_NEW_TASK
		},
	},
	methods: {
		search() {
			this.searchTasks()
			this.searchTeams()
		},
		searchTasks() {
			if (this.searchMode !== SEARCH_MODE_ALL && this.searchMode !== SEARCH_MODE_TASKS) {
				this.foundTasks = []
				return
			}

			if (this.selectedCmd !== null) {
				return
			}

			if (this.taskSearchTimeout !== null) {
				clearTimeout(this.taskSearchTimeout)
				this.taskSearchTimeout = null
			}

			const {text, list, labels} = this.parsedQuery

			const params = {
				s: text,
				filter_by: [],
				filter_value: [],
				filter_comparator: [],
			}

			if (list !== null) {
				const l = this.$store.getters['lists/findListByExactname'](list)
				if (l !== null) {
					params.filter_by.push('list_id')
					params.filter_value.push(l.id)
					params.filter_comparator.push('equals')
				}
			}

			if (labels.length > 0) {
				const labelIds = this.$store.getters['labels/getLabelsByExactTitles'](labels).map(l => l.id)
				if (labelIds.length > 0) {
					params.filter_by.push('labels')
					params.filter_value.push(labelIds.join())
					params.filter_comparator.push('in')
				}
			}

			this.taskSearchTimeout = setTimeout(async () => {
				const r = await this.taskService.getAll({}, params)
				this.foundTasks = r.map(t => {
					t.type = TYPE_TASK
					const list = this.$store.getters['lists/getListById'](t.listId)
					if (list !== null) {
						t.title = `${t.title} (${list.title})`
					}

					return t
				})
			}, 150)
		},
		searchTeams() {
			if (this.searchMode !== SEARCH_MODE_ALL && this.searchMode !== SEARCH_MODE_TEAMS) {
				this.foundTeams = []
				return
			}

			if (this.query === '' || this.selectedCmd !== null) {
				return
			}

			if (this.teamSearchTimeout !== null) {
				clearTimeout(this.teamSearchTimeout)
				this.teamSearchTimeout = null
			}

			const {assignees} = this.parsedQuery

			this.teamSearchTimeout = setTimeout(async () => {
				const teamSearchPromises = assignees.map((t) => this.teamService.getAll({}, {s: t}))
				const teamsResult = await Promise.all(teamSearchPromises)
				this.foundTeams = teamsResult.flatMap(team => {
					team.title = team.name
					return team
				})
			}, 150)
		},
		closeQuickActions() {
			this.$store.commit(QUICK_ACTIONS_ACTIVE, false)
		},
		doAction(type, item) {
			switch (type) {
				case TYPE_LIST:
					this.$router.push({name: 'list.index', params: {listId: item.id}})
					this.closeQuickActions()
					break
				case TYPE_TASK:
					this.$router.push({name: 'task.detail', params: {id: item.id}})
					this.closeQuickActions()
					break
				case TYPE_CMD:
					this.query = ''
					this.selectedCmd = item
					this.$refs.searchInput.focus()
					break
			}
		},
		doCmd() {
			if (this.results.length === 1 && this.results[0].items.length === 1) {
				this.doAction(this.results[0].type, this.results[0].items[0])
				return
			}

			if (this.selectedCmd === null) {
				return
			}

			if (this.query === '') {
				return
			}

			switch (this.selectedCmd.action) {
				case CMD_NEW_TASK:
					this.newTask()
					break
				case CMD_NEW_LIST:
					this.newList()
					break
				case CMD_NEW_NAMESPACE:
					this.newNamespace()
					break
				case CMD_NEW_TEAM:
					this.newTeam()
					break
			}
		},
		async newTask() {
			if (this.currentList === null) {
				return
			}

			const task = await this.$store.dispatch('tasks/createNewTask', {
				title: this.query,
				listId: this.currentList.id,
			})
			this.$message.success({message: this.$t('task.createSuccess')})
			this.$router.push({name: 'task.detail', params: {id: task.id}})
			this.closeQuickActions()
		},
		async newList() {
			if (this.currentList === null) {
				return
			}

			const newList = new ListModel({
				title: this.query,
				namespaceId: this.currentList.namespaceId,
			})
			const list = await this.$store.dispatch('lists/createList', newList)
			this.$message.success({message: this.$t('list.create.createdSuccess')})
			this.$router.push({name: 'list.index', params: {listId: list.id}})
			this.closeQuickActions()
		},
		async newNamespace() {
			const newNamespace = new NamespaceModel({title: this.query})

			await this.$store.dispatch('namespaces/createNamespace', newNamespace)
			this.$message.success({message: this.$t('namespace.create.success')})
			this.closeQuickActions()
		},
		async newTeam() {
			const newTeam = new TeamModel({name: this.query})
			const team = await this.teamService.create(newTeam)
			this.$router.push({
				name: 'teams.edit',
				params: {id: team.id},
			})
			this.$message.success({message: this.$t('team.create.success')})
			this.closeQuickActions()
		},
		select(parentIndex, index) {

			if (index < 0 && parentIndex === 0) {
				this.$refs.searchInput.focus()
				return
			}

			if (index < 0) {
				parentIndex--
				index = this.results[parentIndex].items.length - 1
			}

			let elems = this.$refs[`result-${parentIndex}_${index}`]

			if (this.results[parentIndex].items.length === index) {
				elems = this.$refs[`result-${parentIndex + 1}_0`]
			}

			if (typeof elems === 'undefined' || elems.length === 0) {
				return
			}

			if (Array.isArray(elems)) {
				elems[0].focus()
				return
			}

			elems.focus()
		},
		unselectCmd() {
			if (this.query !== '') {
				return
			}

			this.selectedCmd = null
		},
		reset() {
			this.query = ''
			this.selectedCmd = null
		},
	},
}
</script>

<style lang="scss" scoped>
.quick-actions {
	// FIXME: changed position should be an option of the modal
	:deep(.modal-content) {
		top: 3rem;
		transform: translate(-50%, 0);
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

		.active-cmd {
			font-size: 1.25rem;
			margin-left: .5rem;
		}
	}

	.results {
		text-align: left;
		width: 100%;
		color: $grey-800;

		.result {
			&-title {
				background: $grey-50;
				padding: .5rem;
				display: block;
				font-size: .75rem;
			}

			&-items {
				button {
					font-size: .9rem;
					width: 100%;
					background: transparent;
					text-align: left;
					box-shadow: none;
					border-radius: 0;
					text-transform: none;
					font-family: $family-sans-serif;
					font-weight: normal;
					padding: .5rem .75rem;
					border: none;
					cursor: pointer;

					&:focus, &:hover {
						background: $grey-50;
						box-shadow: none !important;
					}

					&:active {
						background: $grey-100;
					}
				}
			}
		}
	}
}

// HACK:
// FIXME:
.modal-container-smaller :deep(.hint-modal .modal-container) {
	height: calc(100vh - 5rem);
}
</style>