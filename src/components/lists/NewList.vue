<template>
	<div class="fullpage">
		<a class="close" @click="back()">
			<icon :icon="['far', 'times-circle']">
			</icon>
		</a>
		<h3>Create a new list</h3>
		<div class="field is-grouped">
			<p class="control is-expanded" :class="{ 'is-loading': listService.loading}">
				<input v-focus
					class="input"
					:class="{ 'disabled': listService.loading}"
					v-model="list.title"
					type="text"
					placeholder="The list's name goes here..."
					@keyup.esc="back()"
					@keyup.enter="newList()"/>
			</p>
			<p class="control">
				<button class="button is-success noshadow" @click="newList()" :disabled="list.title.length < 3">
						<span class="icon is-small">
							<icon icon="plus"/>
						</span>
					Add
				</button>
			</p>
		</div>
		<p class="help is-danger" v-if="showError && list.title.length < 3">
			Please specify at least three characters.
		</p>
	</div>
</template>

<script>
	import auth from '../../auth'
	import router from '../../router'
	import ListService from '../../services/list'
	import ListModel from '../../models/list'

	export default {
		name: "NewList",
		data() {
			return {
				showError: false,
				list: ListModel,
				listService: ListService,
			}
		},
		beforeMount() {
			// Check if the user is already logged in, if so, redirect him to the homepage
			if (!auth.user.authenticated) {
				router.push({name: 'home'})
			}
		},
		created() {
			this.list = new ListModel()
			this.listService = new ListService()
			this.$parent.setFullPage();
		},
		methods: {
			newList() {
				if (this.list.title.length < 3) {
					this.showError = true
					return
				}
				this.showError = false

				this.list.namespaceId = this.$route.params.id
				this.listService.create(this.list)
					.then(response => {
						this.$parent.loadNamespaces()
						this.success({message: 'The list was successfully created.'}, this)
						router.push({name: 'showList', params: {id: response.id}})
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			back() {
				router.go(-1)
			},
		}
	}
</script>