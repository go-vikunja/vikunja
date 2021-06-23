<template>
	<modal
		@close="$router.back()"
		@submit="archiveList()"
	>
		<span slot="header">{{ list.isArchived ? $t('list.archive.unarchive') : $t('list.archive.archive') }}</span>
		<p slot="text" v-if="list.isArchived">
			{{ $t('list.archive.unarchiveText') }}
		</p>
		<p slot="text" v-else>
			{{ $t('list.archive.archiveText') }}
		</p>
	</modal>
</template>

<script>
import ListService from '@/services/list'

export default {
	name: 'list-setting-archive',
	data() {
		return {
			listService: ListService,
			list: null,
		}
	},
	created() {
		this.listService = new ListService()
		this.list = this.$store.getters['lists/getListById'](this.$route.params.listId)
		this.setTitle(this.$t('list.archive.title', {list: this.list.title}))
	},
	methods: {
		archiveList() {

			this.list.isArchived = !this.list.isArchived

			this.listService.update(this.list)
				.then(r => {
					this.$store.commit('currentList', r)
					this.$store.commit('namespaces/setListInNamespaceById', r)
					this.success({message: this.$t('list.archive.success')})
				})
				.catch(e => {
					this.error(e)
				})
				.finally(() => {
					this.$router.back()
				})
		},
	},
}
</script>
