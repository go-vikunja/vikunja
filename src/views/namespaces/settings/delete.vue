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

<script lang="ts">
import {defineComponent} from 'vue'

export default defineComponent({
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
		async deleteNamespace() {
			await this.$store.dispatch('namespaces/deleteNamespace', this.namespace)
			this.$message.success({message: this.$t('namespace.delete.success')})
			this.$router.push({name: 'home'})
		},
	},
})
</script>
