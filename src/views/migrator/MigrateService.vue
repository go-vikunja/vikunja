<template>
	<migration
			:identifier="identifier"
			:name="name"
	/>
</template>

<script>
	import Migration from '../../components/migrator/migration'
	import router from '../../router'

	export default {
		name: 'migrateService',
		components: {
			Migration,
		},
		data() {
			return {
				name: '',
				identifier: '',
			}
		},
		created() {
			switch (this.$route.params.service) {
				case 'wunderlist':
					this.name = 'Wunderlist'
					this.identifier = 'wunderlist'
					break
				case 'todoist':
					this.name = 'Todoist'
					this.identifier = 'todoist'
					break
				default:
					router.push({name: '404'})
			}
		},
	}
</script>

