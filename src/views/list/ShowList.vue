<template>
	<div
			class="loader-container"
			:class="{ 'is-loading': listService.loading}"
	>
		<div class="switch-view">
			<router-link
					:to="{ name: 'list.list',   params: { listId: listId } }"
					:class="{'is-active': $route.name === 'list.list'}">
				List
			</router-link>
			<router-link
					:to="{ name: 'list.gantt',  params: { listId: listId } }"
					:class="{'is-active': $route.name === 'list.gantt'}">
				Gantt
			</router-link>
			<router-link
					:to="{ name: 'list.table',  params: { listId: listId } }"
					:class="{'is-active': $route.name === 'list.table'}">
				Table
			</router-link>
			<router-link
					:to="{ name: 'list.kanban', params: { listId: listId } }"
					:class="{'is-active': $route.name === 'list.kanban'}">
				Kanban
			</router-link>
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
	import {CURRENT_LIST} from '../../store/mutation-types'
	import {getListView} from '../../helpers/saveListView'

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
		},
		methods: {
			loadList() {

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

				// Don't load the list if we either already loaded it or aren't dealing with a list at all currently
				if (this.$route.params.listId === this.listLoaded || typeof this.$route.params.listId === 'undefined') {
					return
				}

				// Redirect the user to list view by default
				if (
					this.$route.name !== 'list.list' &&
					this.$route.name !== 'list.gantt' &&
					this.$route.name !== 'list.table' &&
					this.$route.name !== 'list.kanban'
				) {

					const savedListView = getListView(this.$route.params.listId)

					router.replace({name: savedListView, params: {id: this.$route.params.listId}})
					return
				}

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
		}
	}
</script>