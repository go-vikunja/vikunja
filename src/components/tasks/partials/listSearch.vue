<template>
	<multiselect
			v-model="list"
			:options="foundLists"
			:multiple="false"
			:searchable="true"
			:loading="listSerivce.loading"
			:internal-search="true"
			@search-change="findLists"
			@select="select"
			placeholder="Type to search for a list..."
			label="title"
			track-by="id"
			:showNoOptions="false"
			class="control is-expanded"
			v-focus
	>
		<template slot="clear" slot-scope="props">
			<div class="multiselect__clear" v-if="list !== null && list.id !== 0" @mousedown.prevent.stop="clearAll(props.search)"></div>
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
	import multiselect from 'vue-multiselect'

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
			multiselect,
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
