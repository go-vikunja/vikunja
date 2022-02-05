<template>
	<modal
		@close="$router.back()"
		@submit="archiveNamespace()"
	>
		<template #header><span>{{ title }}</span></template>

		<template #text>
			<p>
				{{ namespace.isArchived ? $t('namespace.archive.unarchiveText') : $t('namespace.archive.archiveText')}}
			</p>
		</template>
	</modal>
</template>

<script>
import NamespaceService from '@/services/namespace'

export default {
	name: 'namespace-setting-archive',
	data() {
		return {
			namespaceService: new NamespaceService(),
			namespace: null,
			title: '',
		}
	},

	created() {
		this.namespace = this.$store.getters['namespaces/getNamespaceById'](this.$route.params.id)
		this.title = this.namespace.isArchived ?
			this.$t('namespace.archive.titleUnarchive', {namespace: this.namespace.title}) :
			this.$t('namespace.archive.titleArchive', {namespace: this.namespace.title})
		this.setTitle(this.title)
	},

	methods: {
		async archiveNamespace() {
			try {
				const namespace = await this.namespaceService.update({
					...this.namespace,
					isArchived: !this.namespace.isArchived,
				})
				this.$store.commit('namespaces/setNamespaceById', namespace)
				this.$message.success({message: this.$t('namespace.archive.success')})
			} finally {
				this.$router.back()
			}
		},
	},
}
</script>
