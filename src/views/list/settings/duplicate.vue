<template>
	<create-edit
		title="Duplicate this list"
		primary-icon="paste"
		primary-label="Duplicate"
		@primary="duplicateList"
		:loading="listDuplicateService.loading"
	>
		<p>Select a namespace which should hold the duplicated list:</p>
		<namespace-search @selected="selectNamespace"/>
	</create-edit>
</template>

<script>
import ListDuplicateService from '@/services/listDuplicateService'
import NamespaceSearch from '@/components/namespace/namespace-search'
import ListDuplicateModel from '@/models/listDuplicateModel'
import CreateEdit from '@/components/misc/create-edit'

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
		this.setTitle('Duplicate List')
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
					this.$store.commit('lists/addList', r.list)
					this.success({message: 'The list was successfully duplicated.'}, this)
					this.$router.push({name: 'list.index', params: {listId: r.list.id}})
				})
				.catch(e => {
					this.error(e, this)
				})
		},
	},
}
</script>
