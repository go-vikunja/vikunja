<template>
	<div class="loader-container" v-bind:class="{ 'is-loading': namespaceService.loading}">
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Edit Namespace
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form  @submit.prevent="submit()">
						<div class="field">
							<label class="label" for="namespacetext">Namespace Name</label>
							<div class="control">
								<input v-focus :class="{ 'disabled': namespaceService.loading}" :disabled="namespaceService.loading" class="input" type="text" id="namespacetext" placeholder="The namespace text is here..." v-model="namespace.name">
							</div>
						</div>
						<div class="field">
							<label class="label" for="namespacedescription">Description</label>
							<div class="control">
								<textarea :class="{ 'disabled': namespaceService.loading}" :disabled="namespaceService.loading" class="textarea" placeholder="The namespaces description goes here..." id="namespacedescription" v-model="namespace.description"></textarea>
							</div>
						</div>
					</form>

					<div class="columns bigbuttons">
						<div class="column">
							<button @click="submit()" class="button is-primary is-fullwidth" :class="{ 'is-loading': namespaceService.loading}">
								Save
							</button>
						</div>
						<div class="column is-1">
							<button @click="showDeleteModal = true" class="button is-danger is-fullwidth" :class="{ 'is-loading': namespaceService.loading}">
								<span class="icon is-small">
									<icon icon="trash-alt"/>
								</span>
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>

		<component :is="manageUsersComponent" :id="namespace.id" type="namespace" shareType="user" :userIsAdmin="userIsAdmin"></component>
		<component :is="manageTeamsComponent" :id="namespace.id" type="namespace" shareType="team" :userIsAdmin="userIsAdmin"></component>

		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				v-on:submit="deleteNamespace()">
			<span slot="header">Delete the namespace</span>
			<p slot="text">Are you sure you want to delete this namespace and all of its contents?
				<br/>This includes lists & tasks and <b>CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
	import auth from '../../auth'
	import router from '../../router'
	import manageSharing from '../sharing/userTeam'

	import NamespaceService from '../../services/namespace'
	import NamespaceModel from '../../models/namespace'
	
	export default {
		name: "EditNamespace",
		data() {
			return {
				namespaceService: NamespaceService,
				userIsAdmin: false,
				manageUsersComponent: '',
				manageTeamsComponent: '',

				namespace: NamespaceModel,
				showDeleteModal: false,
				user: auth.user,
			}
		},
		components: {
			manageSharing,
		},
		beforeMount() {
			// Check if the user is already logged in, if so, redirect him to the homepage
			if (!auth.user.authenticated) {
				router.push({name: 'home'})
			}

			this.namespace.id = this.$route.params.id
		},
		created() {
			this.namespaceService = new NamespaceService()
			this.namespace = new NamespaceModel()
			this.loadNamespace()
		},
		watch: {
			// call again the method if the route changes
			'$route': 'loadNamespace'
		},
		methods: {
			loadNamespace() {
				let namespace = new NamespaceModel({id: this.$route.params.id})
				this.namespaceService.get(namespace)
					.then(r => {
						this.$set(this, 'namespace', r)
						if (r.owner.id === this.user.infos.id) {
							this.userIsAdmin = true
						}
						// This will trigger the dynamic loading of components once we actually have all the data to pass to them
						this.manageTeamsComponent = 'manageSharing'
						this.manageUsersComponent = 'manageSharing'
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			submit() {
				this.namespaceService.update(this.namespace)
					.then(r => {
						// Update the namespace in the parent
						for (const n in this.$parent.namespaces) {
							if (this.$parent.namespaces[n].id === r.id) {
								r.lists = this.$parent.namespaces[n].lists
								this.$set(this.$parent.namespaces, n, r)
							}
						}
						this.success({message: 'The namespace was successfully updated.'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			deleteNamespace() {
				this.namespaceService.delete(this.namespace)
					.then(() => {
						this.success({message: 'The namespace was successfully deleted.'}, this)
						router.push({name: 'home'})
					})
					.catch(e => {
						this.error(e, this)
					})
			}
		}
	}
</script>