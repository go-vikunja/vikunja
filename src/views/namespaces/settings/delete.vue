<template>
	<modal
		@close="$router.back()"
		@submit="deleteNamespace()"
	>
		<span slot="header">{{ title }}</span>
		<p slot="text">
			{{ $t('namespace.delete.text1') }}<br/>
			{{ $t('namespace.delete.text2') }}
		</p>
	</modal>
</template>

<script>
import NamespaceService from '@/services/namespace'

export default {
	name: 'namespace-setting-delete',
	data() {
		return {
			namespaceService: NamespaceService,
			title: '',
		}
	},
	created() {
		this.namespaceService = new NamespaceService()

		const namespace = this.$store.getters['namespaces/getNamespaceById'](this.$route.params.id)
		this.title = this.$t('namespace.delete.title', {namespace: namespace.title})
		this.setTitle(this.title)
	},
	methods: {
		deleteNamespace() {
			const namespace = this.$store.getters['namespaces/getNamespaceById'](this.$route.params.id)

			this.$store.dispatch('namespaces/deleteNamespace', namespace)
				.then(() => {
					this.success({message: this.$t('namespace.delete.success')})
					this.$router.push({name: 'home'})
				})
				.catch(e => {
					this.error(e)
				})
		},
	},
}
</script>
