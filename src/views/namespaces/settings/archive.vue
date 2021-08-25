<template>
	<modal
		@close="$router.back()"
		@submit="archiveNamespace()"
	>
		<template #header><span>{{ title }}</span></template>
		
		<template #text>
			<p>{{ list.isArchived ? $t('namespace.archive.unarchiveText') : $t('namespace.archive.archiveText') }}</p>
		</template>
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
			title: '',
		}
	},
	created() {
		this.namespaceService = new NamespaceService()
		this.namespace = this.$store.getters['namespaces/getNamespaceById'](this.$route.params.id)
		this.title = this.namespace.isArchived ?
			this.$t('namespace.archive.titleUnarchive', { namespace: this.namespace.title }) :
			this.$t('namespace.archive.titleArchive', { namespace: this.namespace.title })
		this.setTitle(this.title)
	},
	methods: {
		archiveNamespace() {

			this.namespace.isArchived = !this.namespace.isArchived

			this.namespaceService.update(this.namespace)
				.then(r => {
					this.$store.commit('namespaces/setNamespaceById', r)
					this.$message.success({message: this.$t('namespace.archive.success')})
				})
				.catch(e => {
					this.$message.error(e)
				})
				.finally(() => {
					this.$router.back()
				})
		},
	},
}
</script>
