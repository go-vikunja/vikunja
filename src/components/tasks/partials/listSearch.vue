<template>
	<multiselect
		class="control is-expanded"
		v-focus
		:loading="listSerivce.loading"
		:placeholder="$t('list.search')"
		@search="findLists"
		:search-results="foundLists"
		@select="select"
		label="title"
		v-model="list"
		:select-placeholder="$t('list.searchSelect')"
	>
		<template v-slot:searchResult="props">
			<span class="list-namespace-title search-result">{{ namespace(props.option.namespaceId) }} ></span>
			{{ props.option.title }}
		</template>
	</multiselect>
</template>

<script>
import ListService from '../../../services/list'
import ListModel from '../../../models/list'
import Multiselect from '@/components/input/multiselect.vue'

export default {
	name: 'listSearch',
	data() {
		return {
			listSerivce: ListService,
			list: ListModel,
			foundLists: [],
		}
	},
	props: {
		value: {
			required: false,
		},
	},
	components: {
		Multiselect,
	},
	beforeMount() {
		this.listSerivce = new ListService()
		this.list = new ListModel()
	},
	watch: {
		value(newVal) {
			this.list = newVal
		},
	},
	mounted() {
		this.list = this.value
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
			this.list = list
			this.$emit('selected', list)
			this.$emit('input', list)
		},
		namespace(namespaceId) {
			const namespace = this.$store.getters['namespaces/getNamespaceById'](namespaceId)
			if (namespace !== null) {
				return namespace.title
			}
			return this.$t('list.shared')
		},
	},
}
</script>
