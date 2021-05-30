<template>
	<modal v-if="active" class="quick-actions" @close="closeQuickActions">
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

			<div class="has-text-grey-light p-4" v-if="hintText !== ''">
				{{ hintText }}
			</div>

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
import ListService from '@/services/list'
import NamespaceService from '@/services/namespace'
import TeamService from '@/services/team'

import TaskModel from '@/models/task'
import NamespaceModel from '@/models/namespace'
import TeamModel from '@/models/team'

import {CURRENT_LIST, QUICK_ACTIONS_ACTIVE} from '@/store/mutation-types'
import ListModel from '@/models/list'

const TYPE_LIST = 'list'
const TYPE_TASK = 'task'
const TYPE_CMD = 'cmd'
const TYPE_TEAM = 'team'

const CMD_NEW_TASK = 'newTask'
const CMD_NEW_LIST = 'newList'
const CMD_NEW_NAMESPACE = 'newNamespace'
const CMD_NEW_TEAM = 'newTeam'

export default {
	name: 'quick-actions',
	data() {
		return {
			query: '',
			selectedCmd: null,

			foundTasks: [],
			taskSearchTimeout: null,
			taskService: null,

			foundTeams: [],
			teamService: null,

			namespaceService: null,
			listService: null,
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
			const lists = (Object.values(this.$store.state.lists).filter(l => {
				return l.title.toLowerCase().includes(this.query.toLowerCase())
			}) ?? [])

			const cmds = this.availableCmds
				.filter(a => a.title.toLowerCase().includes(this.query.toLowerCase()))

			return [
				{
					type: TYPE_CMD,
					title: 'Commands',
					items: cmds,
				},
				{
					type: TYPE_TASK,
					title: 'Tasks',
					items: this.foundTasks,
				},
				{
					type: TYPE_LIST,
					title: 'Lists',
					items: lists,
				},
				{
					type: TYPE_TEAM,
					title: 'Teams',
					items: this.foundTeams,
				},
			].filter(i => i.items.length > 0)
		},
		nothing() {
			return this.search === '' || Object.keys(this.results).length === 0
		},
		loading() {
			return this.taskService.loading ||
				this.listService.loading ||
				this.namespaceService.loading ||
				this.teamService.loading
		},
		placeholder() {
			if (this.selectedCmd !== null) {
				switch (this.selectedCmd.action) {
					case CMD_NEW_TASK:
						return 'Enter the title of the new task...'
					case CMD_NEW_LIST:
						return 'Enter the title of the new list...'
					case CMD_NEW_NAMESPACE:
						return 'Enter the title of the new namespace...'
					case CMD_NEW_TEAM:
						return 'Enter the name of the new team...'
				}
			}

			return 'Type a command or search...'
		},
		hintText() {
			let namespace

			if (this.selectedCmd !== null && this.currentList !== null) {
				switch (this.selectedCmd.action) {
					case CMD_NEW_TASK:
						return `Create a task in the current list (${this.currentList.title})`
					case CMD_NEW_LIST:
						namespace = this.$store.getters['namespaces/getNamespaceById'](this.currentList.namespaceId)
						return `Create a list in the current namespace (${namespace.title})`
				}
			}

			return ''
		},
		currentList() {
			return Object.keys(this.$store.state[CURRENT_LIST]).length === 0 ? null : this.$store.state[CURRENT_LIST]
		},
		availableCmds() {
			const cmds = []

			if (this.currentList !== null) {
				cmds.push({
					title: 'New task',
					action: CMD_NEW_TASK,
				})
				cmds.push({
					title: 'New list',
					action: CMD_NEW_LIST,
				})
			}
			cmds.push({
				title: 'New namespace',
				action: CMD_NEW_NAMESPACE,
			})
			cmds.push({
				title: 'New Team',
				action: CMD_NEW_TEAM,
			})

			return cmds
		},
	},
	created() {
		this.taskService = new TaskService()
		this.listService = new ListService()
		this.namespaceService = new NamespaceService()
		this.teamService = new TeamService()
	},
	methods: {
		search() {
			this.searchTasks()
		},
		searchTasks() {
			if (this.query === '' || this.selectedCmd !== null) {
				return
			}

			if (this.taskSearchTimeout !== null) {
				clearTimeout(this.taskSearchTimeout)
				this.taskSearchTimeout = null
			}

			this.taskSearchTimeout = setTimeout(() => {
				this.taskService.getAll({}, {s: this.query})
					.then(r => {
						r = r.map(t => {
							t.type = TYPE_TASK
							const list = this.$store.getters['lists/getListById'](t.listId) === null ? null : this.$store.getters['lists/getListById'](t.listId)
							if (list !== null) {
								t.title = `${t.title} (${list.title})`
							}

							return t
						})
						this.$set(this, 'foundTasks', r)
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
		newTask() {
			if (this.currentList === null) {
				return
			}

			const newTask = new TaskModel({
				title: this.query,
				listId: this.currentList.id,
			})
			this.taskService.create(newTask)
				.then(r => {
					this.success({message: 'The task was successfully created.'}, this)
					this.$router.push({name: 'task.detail', params: {id: r.id}})
					this.closeQuickActions()
				})
				.catch((e) => {
					this.error(e, this)
				})
		},
		newList() {
			if (this.currentList === null) {
				return
			}

			const newList = new ListModel({
				title: this.query,
				namespaceId: this.currentList.namespaceId,
			})
			this.listService.create(newList)
				.then(r => {
					this.success({message: 'The list was successfully created.'}, this)
					this.$router.push({name: 'list.index', params: {listId: r.id}})
					this.closeQuickActions()
				})
				.catch((e) => {
					this.error(e, this)
				})
		},
		newNamespace() {
			const newNamespace = new NamespaceModel({title: this.query})
			this.namespaceService.create(newNamespace)
				.then(r => {
					this.$store.commit('namespaces/addNamespace', r)
					this.success({message: 'The namespace was successfully created.'}, this)
					this.$router.back()
					this.closeQuickActions()
				})
				.catch((e) => {
					this.error(e, this)
				})
		},
		newTeam() {
			const newTeam = new TeamModel({name: this.query})
			this.teamService.create(newTeam)
				.then(r => {
					this.$router.push({
						name: 'teams.edit',
						params: {id: r.id},
					})
					this.success({message: 'The team was successfully created.'}, this)
					this.closeQuickActions()
				})
				.catch((e) => {
					this.error(e, this)
				})
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
		}
	},
}
</script>
