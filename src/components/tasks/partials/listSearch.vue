<template>
	<multiselect
		class="control is-expanded"
		v-focus
		:loading="listSerivce.loading"
		placeholder="Type to search for a list..."
		@search="findLists"
		:search-results="foundLists"
		@select="select"
		label="title"
		v-model="list"
		select-placeholder="Click or press enter to select this list"
	>
		<template v-slot:searchResult="props">
			<span class="list-namespace-title">{{ namespace(props.option.namespaceId) }} ></span>
			{{ props.option.title }}
		</template>
	</multiselect>
</template>

<script>
import ListService from '../../../services/list'
import ListModel from '../../../models/list'
import Multiselect from '@/components/input/multiselect'

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
		Multiselect,
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
					this.error(e)
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
