<template>
	<modal
		@close="$router.back()"
		@submit="deleteList()"
	>
		<span slot="header">{{ $t('list.delete.header') }}</span>
		<p slot="text">
			{{ $t('list.delete.text1') }}<br/>
			{{ $t('list.delete.text2') }}
		</p>
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
		this.setTitle(this.$t('list.delete.title', {list: list.title}))
	},
	methods: {
		deleteList() {
			const list = this.$store.getters['lists/getListById'](this.$route.params.listId)

			this.listService.delete(list)
				.then(() => {
					this.$store.commit('namespaces/removeListFromNamespaceById', list)
					this.success({message: this.$t('list.delete.success')})
					this.$router.push({name: 'home'})
				})
				.catch(e => {
					this.error(e)
				})
		},
	},
}
</script>
