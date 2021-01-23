<template>
	<div class="loader-container is-max-width-desktop" v-bind:class="{ 'is-loading': namespaceService.loading}">
		<div class="notification is-warning" v-if="namespace.isArchived">
			This namespace is archived.
			It is not possible to create new lists or edit it.
		</div>
		<card title="Edit Namespace">
			<form @submit.prevent="submit()">
				<div class="field">
					<label class="label" for="namespacetext">Namespace Name</label>
					<div class="control">
						<input
							:class="{ 'disabled': namespaceService.loading}"
							:disabled="namespaceService.loading"
							class="input"
							id="namespacetext"
							placeholder="The namespace text is here..."
							type="text"
							v-focus
							v-model="namespace.title"/>
					</div>
				</div>
				<div class="field">
					<label class="label" for="namespacedescription">Description</label>
					<div class="control">
						<editor
							:class="{ 'disabled': namespaceService.loading}"
							:disabled="namespaceService.loading"
							:preview-is-default="false"
							id="namespacedescription"
							placeholder="The namespaces description goes here..."
							v-if="editorActive"
							v-model="namespace.description"
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

			<div class="field has-addons mt-4">
				<div class="control is-fullwidth">
					<x-button
						@click="submit()"
						:loading="namespaceService.loading"
						class="is-fullwidth"
					>
						Save
					</x-button>
				</div>
				<div class="control">
					<x-button
						@click="showDeleteModal = true"
						:loading="namespaceService.loading"
						class="is-danger"
						icon="trash-alt"
					/>
				</div>
			</div>
		</card>

		<component
			:id="namespace.id"
			:is="manageUsersComponent"
			:userIsAdmin="userIsAdmin"
			shareType="user"
			type="namespace"/>
		<component
			:id="namespace.id"
			:is="manageTeamsComponent"
			:userIsAdmin="userIsAdmin"
			shareType="team"
			type="namespace"/>

		<transition name="modal">
			<modal
				@close="showDeleteModal = false"
				v-if="showDeleteModal"
				@submit="deleteNamespace()">
				<span slot="header">Delete the namespace</span>
				<p slot="text">Are you sure you want to delete this namespace and all of its contents?
					<br/>This includes lists & tasks and <b>CANNOT BE UNDONE!</b></p>
			</modal>
		</transition>
	</div>
</template>

<script>
import router from '../../router'
import manageSharing from '../../components/sharing/userTeam'

import NamespaceService from '../../services/namespace'
import NamespaceModel from '../../models/namespace'
import Fancycheckbox from '../../components/input/fancycheckbox'
import ColorPicker from '../../components/input/colorPicker'
import LoadingComponent from '../../components/misc/loading'
import ErrorComponent from '../../components/misc/error'

export default {
	name: 'EditNamespace',
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
		editor: () => ({
			component: import(/* webpackChunkName: "editor" */ '../../components/input/editor'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
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
		'$route': 'loadNamespace',
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
					this.$store.commit('namespaces/removeNamespaceById', this.namespace.id)
					this.success({message: 'The namespace was successfully deleted.'}, this)
					router.push({name: 'home'})
				})
				.catch(e => {
					this.error(e, this)
				})
		},
	},
}
</script>