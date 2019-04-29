<template>
	<div class="loader-container" :class="{ 'is-loading': listService.loading}">
		<div class="content">
			<router-link :to="{ name: 'editList', params: { id: list.id } }" class="icon settings is-medium">
				<icon icon="cog" size="2x"/>
			</router-link>
			<h1>{{ list.title }}</h1>
			<div class="switch-view">
				<router-link :to="{ name: 'showList', params: { id: list.id } }" :class="{'is-active': $route.params.type !== 'gantt'}">List</router-link>
				<router-link :to="{ name: 'showListWithType', params: { id: list.id, type: 'gantt' } }" :class="{'is-active': $route.params.type === 'gantt'}">Gantt</router-link>
			</div>
		</div>

		<gantt :list="list" v-if="$route.params.type === 'gantt'"/>
		<show-list-task :the-list="list" v-else/>
	</div>
</template>

<script>
	import auth from '../../auth'
	import router from '../../router'
	import message from '../../message'

	import ShowListTask from '../tasks/ShowListTasks'
	import Gantt from '../tasks/Gantt'

	import ListModel from '../../models/list'
	import ListService from '../../services/list'

	export default {
		data() {
			return {
				listID: this.$route.params.id,
				listService: ListService,
				list: ListModel,
			}
		},
		components: {
			Gantt,
			ShowListTask,
		},
		beforeMount() {
			// Check if the user is already logged in, if so, redirect him to the homepage
			if (!auth.user.authenticated) {
				router.push({name: 'home'})
			}

			// If the type is invalid, redirect the user
			if (this.$route.params.type !== 'gantt' && this.$route.params.type !== '') {
				router.push({name: 'showList', params: { id:  this.$route.params.id }})
			}
		},
		created() {
			this.listService = new ListService()
			this.list = new ListModel()
			this.loadList()
		},
		watch: {
			// call again the method if the route changes
			'$route': 'loadList'
		},
		methods: {
			loadList() {
				// We create an extra list object instead of creating it in this.list because that would trigger a ui update which would result in bad ux.
				let list = new ListModel({id: this.$route.params.id})
				this.listService.get(list)
					.then(r => {
						this.$set(this, 'list', r)
					})
					.catch(e => {
						message.error(e, this)
					})
			},
		}
	}
</script>