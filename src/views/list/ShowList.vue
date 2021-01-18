<template>
	<div
		:class="{ 'is-loading': listService.loading}"
		class="loader-container"
	>
		<div class="switch-view-container">
			<div class="switch-view">
				<router-link
					:class="{'is-active': $route.name === 'list.list'}"
					:to="{ name: 'list.list',   params: { listId: listId } }">
					List
				</router-link>
				<router-link
					:class="{'is-active': $route.name === 'list.gantt'}"
					:to="{ name: 'list.gantt',  params: { listId: listId } }">
					Gantt
				</router-link>
				<router-link
					:class="{'is-active': $route.name === 'list.table'}"
					:to="{ name: 'list.table',  params: { listId: listId } }">
					Table
				</router-link>
				<router-link
					:class="{'is-active': $route.name === 'list.kanban'}"
					:to="{ name: 'list.kanban', params: { listId: listId } }">
					Kanban
				</router-link>
			</div>
		</div>
		<div class="notification is-warning" v-if="list.isArchived">
			This list is archived.
			It is not possible to create new or edit tasks or it.
		</div>

		<router-view/>
	</div>
</template>

<script>
import router from '../../router'

import ListModel from '../../models/list'
import ListService from '../../services/list'
import {CURRENT_LIST} from '@/store/mutation-types'
import {getListView} from '@/helpers/saveListView'

export default {
	data() {
		return {
			listService: ListService,
			list: ListModel,
			listLoaded: 0,
		}
	},
	created() {
		this.listService = new ListService()
		this.list = new ListModel()
	},
	mounted() {
		this.loadList()
	},
	watch: {
		// call again the method if the route changes
		'$route.path': 'loadList',
	},
	computed: {
		// Computed property to let "listId" always have a value
		listId() {
			return typeof this.$route.params.listId === 'undefined' ? 0 : this.$route.params.listId
		},
		background() {
			return this.$store.state.background
		},
		currentList() {
			return typeof this.$store.state.currentList === 'undefined' ? {
				id: 0,
				title: '',
			} : this.$store.state.currentList
		},
	},
	methods: {
		replaceListView() {
			const savedListView = getListView(this.$route.params.listId)
			router.replace({name: savedListView, params: {id: this.$route.params.listId}})
			console.debug('Replaced list view with ', savedListView)
			return
		},
		loadList() {
			this.setTitle(this.currentList.title)

			// This invalidates the loaded list at the kanban board which lets it reload its content when
			// switched to it. This ensures updates done to tasks in the gantt or list views are consistently
			// shown in all views while preventing reloads when closing a task popup.
			// We don't do this for the table view because that does not change tasks.
			if (
				this.$route.name === 'list.list' ||
				this.$route.name === 'list.gantt'
			) {
				this.$store.commit('kanban/setListId', 0)
			}

			// When clicking again on a list in the menu, there would be no list view selected which means no list
			// at all. Users will then have to click on the list view menu again which is quite confusing.
			if (this.$route.name === 'list.index') {
				return this.replaceListView()
			}

			// Don't load the list if we either already loaded it or aren't dealing with a list at all currently
			if (
				this.$route.params.listId === this.listLoaded ||
				typeof this.$route.params.listId === 'undefined' ||
				this.$route.params.listId === this.currentList.id ||
				parseInt(this.$route.params.listId) === this.currentList.id
			) {
				return
			}

			// Redirect the user to list view by default
			if (
				this.$route.name !== 'list.list' &&
				this.$route.name !== 'list.gantt' &&
				this.$route.name !== 'list.table' &&
				this.$route.name !== 'list.kanban'
			) {
				return this.replaceListView()
			}

			console.debug(`Loading list, $route.name = ${this.$route.name}, $route.params =`, this.$route.params, `, listLoaded = ${this.listLoaded}, currentList = `, this.currentList)

			// We create an extra list object instead of creating it in this.list because that would trigger a ui update which would result in bad ux.
			let list = new ListModel({id: this.$route.params.listId})
			this.listService.get(list)
				.then(r => {
					this.$set(this, 'list', r)
					this.$store.commit(CURRENT_LIST, r)
				})
				.catch(e => {
					this.error(e, this)
				})
				.finally(() => {
					this.listLoaded = this.$route.params.listId
				})
		},
	},
}
</script>