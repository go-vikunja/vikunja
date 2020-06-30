<template>
	<multiselect
			v-model="namespace"
			:options="namespaces"
			:multiple="false"
			:searchable="true"
			:loading="namespaceService.loading"
			:internal-search="true"
			@search-change="findNamespaces"
			@select="select"
			placeholder="Search for a namespace..."
			label="title"
			track-by="id">
		<template slot="clear" slot-scope="props">
			<div
					class="multiselect__clear" v-if="namespace.id !== 0"
					@mousedown.prevent.stop="clearAll(props.search)"></div>
		</template>
		<span slot="noResult">No namespace found. Consider changing the search query.</span>
	</multiselect>
</template>

<script>
	import multiselect from 'vue-multiselect'

	import NamespaceService from '../../services/namespace'
	import NamespaceModel from '../../models/namespace'

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
			multiselect,
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
