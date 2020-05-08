<template>
	<div class="loader-container" :class="{ 'is-loading': listService.loading}">
		<div class="content">
			<router-link :to="{ name: 'editList', params: { id: list.id } }" class="icon settings is-medium">
				<icon icon="cog" size="2x"/>
			</router-link>
			<h1 :style="{ 'opacity': list.title === '' ? '0': '1' }">{{ list.title === '' ? 'Loading...': list.title}}</h1>
			<div class="notification is-warning" v-if="list.isArchived">
				This list is archived.
				It is not possible to create new or edit tasks or it.
			</div>
			<div class="switch-view">
				<router-link :to="{ name: 'list.list',   params: { id: $route.params.listId } }" :class="{'is-active': $route.name === 'list.list'}">List</router-link>
				<router-link :to="{ name: 'list.gantt',  params: { id: $route.params.listId } }" :class="{'is-active': $route.name === 'list.gantt'}">Gantt</router-link>
				<router-link :to="{ name: 'list.table',  params: { id: $route.params.listId } }" :class="{'is-active': $route.name === 'list.table'}">Table</router-link>
				<router-link :to="{ name: 'list.kanban', params: { id: $route.params.listId } }" :class="{'is-active': $route.name === 'list.kanban'}">Kanban</router-link>
			</div>
		</div>

		<router-view/>
	</div>
</template>

<script>
	import router from '../../router'

	import ListModel from '../../models/list'
	import ListService from '../../services/list'
	import {CURRENT_LIST} from "../../store/mutation-types";

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
		methods: {
			loadList() {

				// Don't load the list if we either already loaded it or aren't dealing with a list at all currently
				if(this.$route.params.listId === this.listLoaded || typeof this.$route.params.listId === 'undefined') {
					return
				}

				// Redirect the user to list view by default
				if (
					this.$route.name !== 'list.list' &&
					this.$route.name !== 'list.gantt' &&
					this.$route.name !== 'list.table' &&
					this.$route.name !== 'list.kanban'
				) {
					router.push({name: 'list.list', params: {id: this.$route.params.listId}})
					return
				}

				this.$store.commit(CURRENT_LIST, Number(this.$route.params.listId))

				// We create an extra list object instead of creating it in this.list because that would trigger a ui update which would result in bad ux.
				let list = new ListModel({id: this.$route.params.listId})
				this.listService.get(list)
					.then(r => {
						this.$set(this, 'list', r)
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