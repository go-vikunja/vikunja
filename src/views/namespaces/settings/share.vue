<template>
	<create-edit
		title="Share this Namespace"
		primary-label=""
	>
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
	</create-edit>
</template>

<script>
import manageSharing from '@/components/sharing/userTeam'
import CreateEdit from '@/components/misc/create-edit'

import NamespaceService from '@/services/namespace'
import NamespaceModel from '@/models/namespace'

export default {
	name: 'namespace-setting-share',
	data() {
		return {
			namespaceService: NamespaceService,
			manageUsersComponent: '',
			manageTeamsComponent: '',

			namespace: NamespaceModel,
		}
	},
	components: {
		CreateEdit,
		manageSharing,
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
			const namespace = new NamespaceModel({id: this.$route.params.id})
			this.namespaceService.get(namespace)
				.then(r => {
					this.$set(this, 'namespace', r)
					// This will trigger the dynamic loading of components once we actually have all the data to pass to them
					this.manageTeamsComponent = 'manageSharing'
					this.manageUsersComponent = 'manageSharing'
					this.setTitle(`Share "${this.namespace.title}"`)
				})
				.catch(e => {
					this.error(e)
				})
		},
	},
}
</script>