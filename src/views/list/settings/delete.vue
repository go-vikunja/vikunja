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
		this.setTitle(this.$t('list.delete.title', {list: this.list.title}))
	},
	computed: {
		list() {
			return this.$store.getters['lists/getListById'](this.$route.params.listId)
		},
	},
	methods: {
		async deleteList() {
			await this.$store.dispatch('lists/deleteList', this.list)
			this.$message.success({message: this.$t('list.delete.success')})
			this.$router.push({name: 'home'})
		},
	},
}
</script>
