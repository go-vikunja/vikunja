<template>
	<div class="loader-container is-max-width-desktop" v-bind:class="{ 'is-loading': namespaceService.loading}">
		<div class="notification is-warning" v-if="namespace.isArchived">
			This namespace is archived.
			It is not possible to create new lists or edit it.
		</div>
		<div class="card">
			<header class="card-header">
				<p class="card-header-title">
					Edit Namespace
				</p>
			</header>
			<div class="card-content">
				<div class="content">
					<form @submit.prevent="submit()">
						<div class="field">
							<label class="label" for="namespacetext">Namespace Name</label>
							<div class="control">
								<input
										v-focus
										:class="{ 'disabled': namespaceService.loading}"
										:disabled="namespaceService.loading"
										class="input"
										type="text"
										id="namespacetext"
										placeholder="The namespace text is here..."
										v-model="namespace.title"/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="namespacedescription">Description</label>
							<div class="control">
								<editor
										:class="{ 'disabled': namespaceService.loading}"
										:disabled="namespaceService.loading"
										placeholder="The namespaces description goes here..."
										id="namespacedescription"
										v-model="namespace.description"
										:preview-is-default="false"
										v-if="editorActive"
								/>
							</div>
						</div>
						<div class="field">
							<label class="label" for="isArchivedCheck">Is Archived</label>
							<div class="control">
								<fancycheckbox
										v-model="namespace.isArchived"
										v-tooltip="'If a namespace is archived, you cannot create new lists or edit it.'">
									This namespace is archived
								</fancycheckbox>
							</div>
						</div>
						<div class="field">
							<label class="label">Color</label>
							<div class="control">
								<color-picker v-model="namespace.hexColor"/>
							</div>
						</div>
					</form>

					<div class="columns bigbuttons">
						<div class="column">
							<button @click="submit()" class="button is-primary is-fullwidth"
									:class="{ 'is-loading': namespaceService.loading}">
								Save
							</button>
						</div>
						<div class="column is-1">
							<button @click="showDeleteModal = true" class="button is-danger is-fullwidth"
									:class="{ 'is-loading': namespaceService.loading}">
								<span class="icon is-small">
									<icon icon="trash-alt"/>
								</span>
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>

		<component
				:is="manageUsersComponent"
				:id="namespace.id"
				type="namespace"
				shareType="user"
				:userIsAdmin="userIsAdmin"/>
		<component
				:is="manageTeamsComponent"
				:id="namespace.id"
				type="namespace"
				shareType="team"
				:userIsAdmin="userIsAdmin"/>

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
	import router from '../../router'
	import manageSharing from '../../components/sharing/userTeam'

	import NamespaceService from '../../services/namespace'
	import NamespaceModel from '../../models/namespace'
	import Fancycheckbox from '../../components/input/fancycheckbox'
	import ColorPicker from '../../components/input/colorPicker'

	export default {
		name: "EditNamespace",
		data() {
			return {
				namespaceService: NamespaceService,
				manageUsersComponent: '',
				manageTeamsComponent: '',

				namespace: NamespaceModel,
				showDeleteModal: false,
				editorActive: false,
			}
		},
		components: {
			ColorPicker,
			Fancycheckbox,
			manageSharing,
			editor: () => import(/* webpackPrefetch: true */ '../../components/input/editor'),
		},
		beforeMount() {
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
		computed: {
			userIsAdmin() {
				return this.namespace.owner && this.namespace.owner.id === this.$store.state.auth.info.id
			},
		},
		methods: {
			loadNamespace() {
				// This makes the editor trigger its mounted function again which makes it forget every input
				// it currently has in its textarea. This is a counter-hack to a hack inside of vue-easymde
				// which made it impossible to detect change from the outside. Therefore the component would
				// not update if new content from the outside was made available.
				// See https://github.com/NikulinIlya/vue-easymde/issues/3
				this.editorActive = false
				this.$nextTick(() => this.editorActive = true)

				let namespace = new NamespaceModel({id: this.$route.params.id})
				this.namespaceService.get(namespace)
					.then(r => {
						this.$set(this, 'namespace', r)
						// This will trigger the dynamic loading of components once we actually have all the data to pass to them
						this.manageTeamsComponent = 'manageSharing'
						this.manageUsersComponent = 'manageSharing'
						this.setTitle(`Edit ${this.namespace.title}`)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
			submit() {
				this.namespaceService.update(this.namespace)
					.then(r => {
						// Update the namespace in the parent
						this.$store.commit('namespaces/setNamespaceById', r)
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