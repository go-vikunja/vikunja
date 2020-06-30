<template>
	<div class="loader-container edit-list" :class="{ 'is-loading': listService.loading}">
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
					<form @submit.prevent="submit()">
						<div class="field">
							<label class="label" for="listtext">List Name</label>
							<div class="control">
								<input
										v-focus
										:class="{ 'disabled': listService.loading}"
										:disabled="listService.loading"
										class="input"
										type="text"
										id="listtext"
										placeholder="The list title goes here..."
										@keyup.enter="submit"
										v-model="list.title"/>
							</div>
						</div>
						<div class="field">
							<label
									class="label"
									for="listtext"
									v-tooltip="'The list identifier can be used to uniquely identify a task across lists. You can set it to empty to disable it.'">
								List Identifier
							</label>
							<div class="control">
								<input
										v-focus
										:class="{ 'disabled': listService.loading}"
										:disabled="listService.loading"
										class="input"
										type="text"
										id="listtext"
										placeholder="The list identifier goes here..."
										@keyup.enter="submit"
										v-model="list.identifier"/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="listdescription">Description</label>
							<div class="control">
								<textarea
										:class="{ 'disabled': listService.loading}"
										:disabled="listService.loading"
										class="textarea"
										placeholder="The lists description goes here..."
										id="listdescription"
										v-model="list.description"></textarea>
							</div>
						</div>
						<div class="field">
							<label class="label" for="isArchivedCheck">Is Archived</label>
							<div class="control">
								<fancycheckbox
										v-model="list.isArchived"
										v-tooltip="'If a list is archived, you cannot create new tasks or edit the list or existing tasks.'">
									This list is archived
								</fancycheckbox>
							</div>
						</div>
						<div class="field">
							<label class="label">Color</label>
							<div class="control">
								<color-picker v-model="list.hexColor"/>
							</div>
						</div>
					</form>

					<div class="columns bigbuttons">
						<div class="column">
							<button @click="submit()" class="button is-primary is-fullwidth"
									:class="{ 'is-loading': listService.loading}">
								Save
							</button>
						</div>
						<div class="column is-1">
							<button @click="showDeleteModal = true" class="button is-danger is-fullwidth"
									:class="{ 'is-loading': listService.loading}">
								<span class="icon is-small">
									<icon icon="trash-alt"/>
								</span>
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>

		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Duplicate this list
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<p>Select a namespace which should hold the duplicated list:</p>
					<div class="field is-grouped">
						<p class="control is-expanded">
							<namespace-search @selected="selectNamespace"/>
						</p>
						<p class="control">
							<button type="submit" class="button is-success" @click="duplicateList">
								<span class="icon is-small">
									<icon icon="plus"/>
								</span>
								Add
							</button>
						</p>
					</div>
				</div>
			</div>
		</div>

		<background :list-id="$route.params.id"/>

		<component
				:is="manageUsersComponent"
				:id="list.id"
				type="list"
				shareType="user"
				:userIsAdmin="userIsAdmin"/>
		<component
				:is="manageTeamsComponent"
				:id="list.id"
				type="list"
				shareType="team"
				:userIsAdmin="userIsAdmin"/>

		<link-sharing :list-id="$route.params.id" v-if="linkSharingEnabled"/>

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
	import router from '../../router'
	import manageSharing from '../../components/sharing/userTeam'
	import LinkSharing from '../../components/sharing/linkSharing'

	import ListModel from '../../models/list'
	import ListService from '../../services/list'
	import Fancycheckbox from '../../components/input/fancycheckbox'
	import Background from '../../components/list/partials/background-settings'
	import {CURRENT_LIST} from '../../store/mutation-types'
	import ColorPicker from '../../components/input/colorPicker'
	import NamespaceSearch from '../../components/namespace/namespace-search'
	import ListDuplicateService from '../../services/listDuplicateService'
	import ListDuplicateModel from '../../models/listDuplicateModel'

	export default {
		name: 'EditList',
		data() {
			return {
				list: ListModel,
				listService: ListService,

				showDeleteModal: false,

				manageUsersComponent: '',
				manageTeamsComponent: '',

				listDuplicateService: ListDuplicateService,
				selectedNamespace: null,
			}
		},
		components: {
			NamespaceSearch,
			ColorPicker,
			Background,
			Fancycheckbox,
			LinkSharing,
			manageSharing,
		},
		created() {
			this.listService = new ListService()
			this.listDuplicateService = new ListDuplicateService()
			this.loadList()
		},
		watch: {
			// call again the method if the route changes
			'$route': 'loadList'
		},
		computed: {
			linkSharingEnabled() {
				return this.$store.state.config.linkSharingEnabled
			},
			userIsAdmin() {
				return this.list.owner && this.list.owner.id === this.$store.state.auth.info.id
			},
		},
		methods: {
			loadList() {
				let list = new ListModel({id: this.$route.params.id})
				this.listService.get(list)
					.then(r => {
						this.$set(this, 'list', r)
						this.$store.commit(CURRENT_LIST, r)
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
						this.$store.commit('namespaces/setListInNamespaceById', r)
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
			selectNamespace(namespace) {
				this.selectedNamespace = namespace
			},
			duplicateList() {
				const listDuplicate = new ListDuplicateModel({
					listId: this.list.id,
					namespaceId: this.selectedNamespace.id,
				})
				this.listDuplicateService.create(listDuplicate)
					.then(r => {
						this.$store.commit('namespaces/addListToNamespace', r.list)
						this.success({message: 'The list was successfully duplicated.'}, this)
						router.push({name: 'list.index', params: {listId: r.list.id}})
					})
					.catch(e => {
						this.error(e, this)
					})
			},
		}
	}
</script>
