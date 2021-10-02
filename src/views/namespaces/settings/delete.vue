<template>
	<modal
		@close="$router.back()"
		@submit="deleteNamespace()"
	>
		<template #header><span>{{ title }}</span></template>
		
		<template #text>
			<p>{{ $t('namespace.delete.text1') }}<br/>
			{{ $t('namespace.delete.text2') }}</p>
		</template>
	</modal>
</template>

<script>
export default {
	name: 'namespace-setting-delete',
	computed: {
		namespace() {
			return this.$store.getters['namespaces/getNamespaceById'](this.$route.params.id)
		},
		title() {
			if (!this.namespace) {
				return
			}
			return this.$t('namespace.delete.title', {namespace: this.namespace.title})
		},
	},
	watch: {
		title: {
			handler(title) {
				this.setTitle(title)
			},
			immediate: true,
		},
	},
	methods: {
		deleteNamespace() {
			this.$store.dispatch('namespaces/deleteNamespace', this.namespace)
				.then(() => {
					this.$message.success({message: this.$t('namespace.delete.success')})
					this.$router.push({name: 'home'})
				})
				.catch(e => {
					this.$message.error(e)
				})
		},
	},
}
</script>
