<template>
	<modal
		@close="$router.back()"
		@submit="archiveNamespace()"
	>
		<span slot="header">{{ namespace.isArchived ? 'Un-' : '' }}Archive this namespace</span>
		<p slot="text" v-if="namespace.isArchived">
			You will be able to create new lists or edit it.
		</p>
		<p slot="text" v-else>
			You won't be able to edit this namespace or create new list until you un-archive it.<br/>
			This will also archive all lists in this namespace.
		</p>
	</modal>
</template>

<script>
import NamespaceService from '@/services/namespace'

export default {
	name: 'namespace-setting-archive',
	data() {
		return {
			namespaceService: NamespaceService,
			namespace: null,
		}
	},
	created() {
		this.namespaceService = new NamespaceService()
		this.namespace = this.$store.getters['namespaces/getNamespaceById'](this.$route.params.id)
		this.setTitle(`Archive "${this.namespace.title}"`)
	},
	methods: {
		archiveNamespace() {

			this.namespace.isArchived = !this.namespace.isArchived

			this.namespaceService.update(this.namespace)
				.then(r => {
					this.$store.commit('namespaces/setNamespaceById', r)
					this.success({message: 'The namespace was successfully archived.'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
				.finally(() => {
					this.$router.back()
				})
		},
	},
}
</script>
