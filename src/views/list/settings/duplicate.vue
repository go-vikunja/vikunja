<template>
	<create-edit
		:title="$t('list.duplicate.title')"
		primary-icon="paste"
		:primary-label="$t('list.duplicate.label')"
		@primary="duplicateList"
		:loading="listDuplicateService.loading"
	>
		<p>
			{{ $t('list.duplicate.text') }}
		</p>
		<namespace-search @selected="selectNamespace"/>
	</create-edit>
</template>

<script>
import ListDuplicateService from '@/services/listDuplicateService'
import NamespaceSearch from '@/components/namespace/namespace-search.vue'
import ListDuplicateModel from '@/models/listDuplicateModel'
import CreateEdit from '@/components/misc/create-edit.vue'

export default {
	name: 'list-setting-duplicate',
	data() {
		return {
			listDuplicateService: ListDuplicateService,
			selectedNamespace: null,
		}
	},
	components: {
		CreateEdit,
		NamespaceSearch,
	},
	created() {
		this.listDuplicateService = new ListDuplicateService()
		this.setTitle(this.$t('list.duplicate.title'))
	},
	methods: {
		selectNamespace(namespace) {
			this.selectedNamespace = namespace
		},
		duplicateList() {
			const listDuplicate = new ListDuplicateModel({
				listId: this.$route.params.listId,
				namespaceId: this.selectedNamespace.id,
			})
			this.listDuplicateService.create(listDuplicate)
				.then(r => {
					this.$store.commit('namespaces/addListToNamespace', r.list)
					this.$store.commit('lists/setList', r.list)
					this.$message.success({message: this.$t('list.duplicate.success')})
					this.$router.push({name: 'list.index', params: {listId: r.list.id}})
				})
				.catch(e => {
					this.$message.error(e)
				})
		},
	},
}
</script>
