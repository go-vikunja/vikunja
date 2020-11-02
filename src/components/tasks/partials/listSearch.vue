<template>
	<multiselect
		:internal-search="true"
		:loading="listSerivce.loading"
		:multiple="false"
		:options="foundLists"
		:searchable="true"
		:showNoOptions="false"
		@search-change="findLists"
		@select="select"
		class="control is-expanded"
		label="title"
		placeholder="Type to search for a list..."
		track-by="id"
		v-focus
		v-model="list"
	>
		<template slot="clear" slot-scope="props">
			<div
				@mousedown.prevent.stop="clearAll(props.search)"
				class="multiselect__clear"
				v-if="list !== null && list.id !== 0"></div>
		</template>
		<template slot="option" slot-scope="props">
			<span class="list-namespace-title">{{ namespace(props.option.namespaceId) }} ></span>
			{{ props.option.title }}
		</template>
		<span slot="noResult">No list found. Consider changing the search query.</span>
	</multiselect>
</template>

<script>
import ListService from '../../../services/list'
import ListModel from '../../../models/list'
import LoadingComponent from '../../misc/loading'
import ErrorComponent from '../../misc/error'

export default {
	name: 'listSearch',
	data() {
		return {
			listSerivce: ListService,
			list: ListModel,
			foundLists: [],
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
	beforeMount() {
		this.listSerivce = new ListService()
		this.list = new ListModel()
	},
	methods: {
		findLists(query) {
			if (query === '') {
				this.clearAll()
				return
			}

			this.listSerivce.getAll({}, {s: query})
				.then(response => {
					this.$set(this, 'foundLists', response)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		clearAll() {
			this.$set(this, 'foundLists', [])
		},
		select(list) {
			this.$emit('selected', list)
		},
		namespace(namespaceId) {
			const namespace = this.$store.getters['namespaces/getNamespaceById'](namespaceId)
			if (namespace !== null) {
				return namespace.title
			}
			return 'Shared Lists'
		},
	},
}
</script>
