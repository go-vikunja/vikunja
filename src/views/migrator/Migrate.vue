<template>
	<div class="content">
		<h1>{{ $t('migrate.title') }}</h1>
		<p>{{ $t('migrate.description') }}</p>
		<div class="migration-services-overview">
			<router-link
				:key="m.identifier"
				:to="{name: 'migrate.service', params: {service: m.identifier}}"
				v-for="m in availableMigrators">
				<img :alt="m.name" :src="`/images/migration/${m.identifier}.png`"/>
				{{ m.name }}
			</router-link>
		</div>
	</div>
</template>

<script>
import {getMigratorFromSlug} from '../../helpers/migrator'

export default {
	name: 'migrate.service',
	mounted() {
		this.setTitle(this.$t('migrate.title'))
	},
	computed: {
		availableMigrators() {
			return this.$store.state.config.availableMigrators.map(getMigratorFromSlug)
		},
	},
}
</script>
