<template>
	<create-edit
		:title="title"
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
import manageSharing from '@/components/sharing/userTeam.vue'
import CreateEdit from '@/components/misc/create-edit.vue'

import NamespaceService from '@/services/namespace'
import NamespaceModel from '@/models/namespace'

export default {
	name: 'namespace-setting-share',
	data() {
		return {
			namespaceService: new NamespaceService(),
			namespace: new NamespaceModel(),
			manageUsersComponent: '',
			manageTeamsComponent: '',
			title: '',
		}
	},
	components: {
		CreateEdit,
		manageSharing,
	},
	beforeMount() {
		this.namespace.id = this.$route.params.id
	},
	watch: {
		// call again the method if the route changes
		'$route': {
			handler: 'loadNamespace',
			deep: true,
			immediate: true,
		},
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
					this.namespace = r
					// This will trigger the dynamic loading of components once we actually have all the data to pass to them
					this.manageTeamsComponent = 'manageSharing'
					this.manageUsersComponent = 'manageSharing'
					this.title = this.$t('namespace.share.title', { namespace: this.namespace.title })
					this.setTitle(this.title)
				})
		},
	},
}
</script>