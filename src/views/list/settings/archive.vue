<template>
	<modal
		@close="$router.back()"
		@submit="archiveList()"
	>
		<span slot="header">{{ list.isArchived ? 'Un-' : '' }}Archive this list</span>
		<p slot="text" v-if="list.isArchived">
			You will be able to create new tasks or edit it.
		</p>
		<p slot="text" v-else>
			You won't be able to edit this list or create new tasks until you un-archive it.
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
		this.setTitle(`Archive "${this.list.title}"`)
	},
	methods: {
		archiveList() {

			this.list.isArchived = !this.list.isArchived

			this.listService.update(this.list)
				.then(r => {
					this.$store.commit('currentList', r)
					this.$store.commit('namespaces/setListInNamespaceById', r)
					this.success({message: 'The list was successfully archived.'})
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
