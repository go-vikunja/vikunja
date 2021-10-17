<template>
	<migration
		:identifier="identifier"
		:name="name"
		:is-file-migrator="isFileMigrator"
	/>
</template>

<script>
import Migration from '../../components/migrator/migration'
import {getMigratorFromSlug} from '../../helpers/migrator'

export default {
	name: 'migrateService',
	components: {
		Migration,
	},
	data() {
		return {
			name: '',
			identifier: '',
			isFileMigrator: false,
		}
	},
	mounted() {
		this.setTitle(this.$t('migrate.titleService', {name: this.name}))
	},
	created() {
		try {
			const {name, identifier, isFileMigrator} = getMigratorFromSlug(this.$route.params.service)
			this.name = name
			this.identifier = identifier
			this.isFileMigrator = isFileMigrator
		} catch (e) {
			this.$router.push({name: 'not-found'})
		}
	},
}
</script>

