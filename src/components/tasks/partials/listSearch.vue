<template>
	<multiselect
		class="control is-expanded"
		:loading="listSerivce.loading"
		:placeholder="$t('list.search')"
		@search="findLists"
		:search-results="foundLists"
		@select="select"
		label="title"
		v-model="list"
		:select-placeholder="$t('list.searchSelect')"
	>
		<template #searchResult="props">
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
			listSerivce: new ListService(),
			list: new ListModel(),
			foundLists: [],
		}
	},
	props: {
		modelValue: {
			required: false,
		},
	},
	emits: ['update:modelValue', 'selected'],
	components: {
		Multiselect,
	},
	watch: {
		modelValue: {
			handler(value) {
				this.list = value
			},
			immeditate: true,
			deep: true,
		},
	},
	methods: {
		async findLists(query) {
			if (query === '') {
				this.clearAll()
				return
			}

			this.foundLists = await this.listSerivce.getAll({}, {s: query})
		},

		clearAll() {
			this.foundLists = []
		},

		select(list) {
			this.list = list
			this.$emit('selected', list)
			this.$emit('update:modelValue', list)
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
