<template>
	<modal
		@close="$router.back()"
		@submit="deleteNamespace()"
	>
		<span slot="header">Delete this namespace</span>
		<p slot="text">Are you sure you want to delete this namespace and all of its contents?
			<br/>This includes all tasks and <b>CANNOT BE UNDONE!</b></p>
	</modal>
</template>

<script>
import NamespaceService from '@/services/namespace'

export default {
	name: 'namespace-setting-delete',
	data() {
		return {
			namespaceService: NamespaceService,
		}
	},
	created() {
		this.namespaceService = new NamespaceService()

		const namespace = this.$store.getters['namespaces/getNamespaceById'](this.$route.params.id)
		this.setTitle(`Delete "${namespace.title}"`)
	},
	methods: {
		deleteNamespace() {
			const namespace = this.$store.getters['namespaces/getNamespaceById'](this.$route.params.id)

			this.$store.dispatch('namespaces/deleteNamespace', namespace)
				.then(() => {
					this.success({message: 'The namespace was successfully deleted.'})
					this.$router.push({name: 'home'})
				})
				.catch(e => {
					this.error(e)
				})
		},
	},
}
</script>
