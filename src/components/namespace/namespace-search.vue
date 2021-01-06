<template>
	<multiselect
		:loading="namespaceService.loading"
		placeholder="Search for a namespace..."
		@search="findNamespaces"
		:search-results="namespaces"
		@select="select"
		label="title"
		v-model="namespace"
	/>
</template>

<script>
import NamespaceService from '../../services/namespace'
import NamespaceModel from '../../models/namespace'

import Multiselect from '@/components/input/multiselect'

export default {
	name: 'namespace-search',
	data() {
		return {
			namespaceService: NamespaceService,
			namespace: NamespaceModel,
			namespaces: [],
		}
	},
	components: {
		Multiselect,
	},
	created() {
		this.namespaceService = new NamespaceService()
	},
	methods: {
		findNamespaces(query) {
			if (query === '') {
				this.clearAll()
				return
			}

			this.namespaceService.getAll({}, {s: query})
				.then(response => {
					this.$set(this, 'namespaces', response)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		clearAll() {
			this.$set(this, 'namespaces', [])
		},
		select(namespace) {
			this.$emit('selected', namespace)
		},
	},
}
</script>
