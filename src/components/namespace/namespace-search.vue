<template>
	<multiselect
		:internal-search="true"
		:loading="namespaceService.loading"
		:multiple="false"
		:options="namespaces"
		:searchable="true"
		:showNoOptions="false"
		@search-change="findNamespaces"
		@select="select"
		label="title"
		placeholder="Search for a namespace..."
		track-by="id"
		v-model="namespace">
		<template slot="clear" slot-scope="props">
			<div
				@mousedown.prevent.stop="clearAll(props.search)" class="multiselect__clear"
				v-if="namespace.id !== 0"></div>
		</template>
		<span slot="noResult">No namespace found. Consider changing the search query.</span>
	</multiselect>
</template>

<script>
import NamespaceService from '../../services/namespace'
import NamespaceModel from '../../models/namespace'
import LoadingComponent from '../misc/loading'
import ErrorComponent from '../misc/error'

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
		multiselect: () => ({
			component: import(/* webpackChunkName: "multiselect" */ 'vue-multiselect'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
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
