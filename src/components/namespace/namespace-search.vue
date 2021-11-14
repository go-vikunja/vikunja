<template>
	<multiselect
		:placeholder="$t('namespace.search')"
		@search="findNamespaces"
		:search-results="namespaces"
		@select="select"
		label="title"
		:search-delay="10"
	/>
</template>

<script>
import Multiselect from '@/components/input/multiselect.vue'

export default {
	name: 'namespace-search',
	emits: ['selected'],
	data() {
		return {
			query: '',
		}
	},
	components: {
		Multiselect,
	},
	computed: {
		namespaces() {
			return this.$store.getters['namespaces/searchNamespace'](this.query)
		},
	},
	methods: {
		findNamespaces(query) {
			this.query = query
		},
		select(namespace) {
			this.$emit('selected', namespace)
		},
	},
}
</script>
