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

<script lang="ts">
import {defineComponent} from 'vue'

import ListDuplicateService from '@/services/listDuplicateService'
import NamespaceSearch from '@/components/namespace/namespace-search.vue'
import ListDuplicateModel from '@/models/listDuplicateModel'
import CreateEdit from '@/components/misc/create-edit.vue'

export default defineComponent({
	name: 'list-setting-duplicate',
	data() {
		return {
			listDuplicateService: new ListDuplicateService(),
			selectedNamespace: null,
		}
	},
	components: {
		CreateEdit,
		NamespaceSearch,
	},
	created() {
		this.setTitle(this.$t('list.duplicate.title'))
	},
	methods: {
		selectNamespace(namespace) {
			this.selectedNamespace = namespace
		},

		async duplicateList() {
			const listDuplicate = new ListDuplicateModel({
				listId: this.$route.params.listId,
				namespaceId: this.selectedNamespace.id,
			})
			const duplicate = await this.listDuplicateService.create(listDuplicate)
			this.$store.commit('namespaces/addListToNamespace', duplicate.list)
			this.$store.commit('lists/setList', duplicate.list)
			this.$message.success({message: this.$t('list.duplicate.success')})
			this.$router.push({name: 'list.index', params: {listId: duplicate.list.id}})
		},
	},
})
</script>
