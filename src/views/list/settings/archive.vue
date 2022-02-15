<template>
	<modal
		@close="$router.back()"
		@submit="archiveList()"
	>
		<template #header><span>{{ list.isArchived ? $t('list.archive.unarchive') : $t('list.archive.archive') }}</span></template>
		
		<template #text>
			<p>{{ list.isArchived ? $t('list.archive.unarchiveText') : $t('list.archive.archiveText') }}</p>
		</template>
	</modal>
</template>

<script lang="ts">
import ListService from '@/services/list'

export default {
	name: 'list-setting-archive',
	data() {
		return {
			listService: new ListService(),
		}
	},
	created() {
		this.setTitle(this.$t('list.archive.title', {list: this.list.title}))
	},
	computed: {
		list() {
			return this.$store.getters['lists/getListById'](this.$route.params.listId)
		},
	},
	methods: {
		async archiveList() {
			const newList = {
				...this.list,
				isArchived: !this.list.isArchived,
			}

			try {
				const list = await this.listService.update(newList)
				this.$store.commit('currentList', list)
				this.$store.commit('namespaces/setListInNamespaceById', list)
				this.$message.success({message: this.$t('list.archive.success')})
			} finally {
				this.$router.back()
			}
		},
	},
}
</script>
