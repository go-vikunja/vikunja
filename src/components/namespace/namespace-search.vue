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
			if (this.query === '') {
				return []
			}
			
			return this.$store.state.namespaces.namespaces.filter(n => {
				return !n.isArchived && 
					n.id > 0 &&
					n.title.toLowerCase().includes(this.query.toLowerCase())
			})
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
