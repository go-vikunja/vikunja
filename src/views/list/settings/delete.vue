<template>
	<modal
		@close="$router.back()"
		@submit="deleteList()"
	>
		<span slot="header">Delete this list</span>
		<p slot="text">Are you sure you want to delete this list and all of its contents?
			<br/>This includes all tasks and <b>CANNOT BE UNDONE!</b></p>
	</modal>
</template>

<script>
import ListService from '@/services/list'

export default {
	name: 'list-setting-delete',
	data() {
		return {
			listService: ListService,
		}
	},
	created() {
		this.listService = new ListService()
		const list = this.$store.getters['lists/getListById'](this.$route.params.listId)
		this.setTitle(`Delete "${list.title}"`)
	},
	methods: {
		deleteList() {
			const list = this.$store.getters['lists/getListById'](this.$route.params.listId)

			this.listService.delete(list)
				.then(() => {
					this.$store.commit('namespaces/removeListFromNamespaceById', list)
					this.success({message: 'The list was successfully deleted.'})
					this.$router.push({name: 'home'})
				})
				.catch(e => {
					this.error(e)
				})
		},
	},
}
</script>
