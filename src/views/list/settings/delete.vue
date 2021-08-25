<template>
	<modal
		@close="$router.back()"
		@submit="deleteList()"
	>
		<template #header><span>{{ $t('list.delete.header') }}</span></template>
		
		<template #text>
			<p>{{ $t('list.delete.text1') }}<br/>
			{{ $t('list.delete.text2') }}</p>
		</template>
	</modal>
</template>

<script>
export default {
	name: 'list-setting-delete',
	created() {
		const list = this.$store.getters['lists/getListById'](this.$route.params.listId)
		this.setTitle(this.$t('list.delete.title', {list: list.title}))
	},
	methods: {
		deleteList() {
			const list = this.$store.getters['lists/getListById'](this.$route.params.listId)

			this.$store.dispatch('lists/deleteList', list)
				.then(() => {
					this.$message.success({message: this.$t('list.delete.success')})
					this.$router.push({name: 'home'})
				})
				.catch(e => {
					this.$message.error(e)
				})
		},
	},
}
</script>
