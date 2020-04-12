<template>
	<div class="loader-container" :class="{ 'is-loading': listService.loading}">
		<div class="notification is-warning" v-if="list.isArchived">
			This list is archived.
			It is not possible to create new or edit tasks or it.
		</div>
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Edit List
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form  @submit.prevent="submit()">
						<div class="field">
							<label class="label" for="listtext">List Name</label>
							<div class="control">
								<input v-focus :class="{ 'disabled': listService.loading}" :disabled="listService.loading" class="input" type="text" id="listtext" placeholder="The list title goes here..." v-model="list.title">
							</div>
						</div>
						<div class="field">
							<label class="label" for="listdescription">Description</label>
							<div class="control">
								<textarea :class="{ 'disabled': listService.loading}" :disabled="listService.loading" class="textarea" placeholder="The lists description goes here..." id="listdescription" v-model="list.description"></textarea>
							</div>
						</div>
						<div class="field">
							<label class="label" for="isArchivedCheck">Is Archived</label>
							<div class="control">
								<fancycheckbox v-model="list.isArchived" v-tooltip="'If a list is archived, you cannot create new tasks or edit the list or existing tasks.'">
									This list is archived
								</fancycheckbox>
							</div>
						</div>
						<div class="field">
							<label class="label">Color</label>
							<div class="control">
								<verte
										v-model="list.hexColor"
										menuPosition="top"
										picker="square"
										model="hex"
										:enableAlpha="false"
										:rgbSliders="true"/>
							</div>
						</div>
					</form>

					<div class="columns bigbuttons">
						<div class="column">
							<button @click="submit()" class="button is-primary is-fullwidth" :class="{ 'is-loading': listService.loading}">
								Save
							</button>
						</div>
						<div class="column is-1">
							<button @click="showDeleteModal = true" class="button is-danger is-fullwidth" :class="{ 'is-loading': listService.loading}">
								<span class="icon is-small">
									<icon icon="trash-alt"/>
								</span>
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>

		<component :is="manageUsersComponent" :id="list.id" type="list" shareType="user" :userIsAdmin="userIsAdmin"></component>
		<component :is="manageTeamsComponent" :id="list.id" type="list" shareType="team" :userIsAdmin="userIsAdmin"></component>

		<link-sharing :list-i-d="$route.params.id"/>

		<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				@submit="deleteList()">
			<span slot="header">Delete the list</span>
			<p slot="text">Are you sure you want to delete this list and all of its contents?
				<br/>This includes all tasks and <b>CANNOT BE UNDONE!</b></p>
		</modal>
	</div>
</template>

<script>
	import verte from 'verte'
	import 'verte/dist/verte.css'

	import auth from '../../auth'
	import router from '../../router'
	import manageSharing from '../sharing/userTeam'
	import LinkSharing from '../sharing/linkSharing'

	import ListModel from '../../models/list'
	import ListService from '../../services/list'
	import Fancycheckbox from '../global/fancycheckbox'

	export default {
		name: "EditList",
		data() {
			return {
				list: ListModel,
				listService: ListService,

				showDeleteModal: false,
				user: auth.user,
				userIsAdmin: false, // FIXME: we should be able to know somehow if the user is admin, not only based on if he's the owner

				manageUsersComponent: '',
				manageTeamsComponent: '',
			}
		},
		components: {
			Fancycheckbox,
			LinkSharing,
			manageSharing,
			verte,
		},
		beforeMount() {
			// Check if the user is already logged in, if so, redirect him to the homepage
			if (!auth.user.authenticated) {
				router.push({name: 'home'})
			}
		},
		created() {
			this.listService = new ListService()
			this.loadList()
		},
		watch: {
			// call again the method if the route changes
			'$route': 'loadList'
		},
		methods: {
			loadList() {
				let list = new ListModel({id: this.$route.params.id})
				this.listService.get(list)
					.then(r => {
						this.$set(this, 'list', r)
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
				this.listService.update(this.list)
					.then(r => {
						// Update the list in the parent
						for (const n in this.$parent.namespaces) {
							let lists = this.$parent.namespaces[n].lists
							for (const l in lists) {
								if (lists[l].id === r.id) {
									this.$set(this.$parent.namespaces[n].lists, l, r)
								}
							}
						}
						this.success({message: 'The list was successfully updated.'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			deleteList() {
				this.listService.delete(this.list)
					.then(() => {
						this.success({message: 'The list was successfully deleted.'}, this)
						router.push({name: 'home'})
					})
					.catch(e => {
						this.error(e, this)
					})
			},
		}
	}
</script>
