<template>
	<div class="content">
		<h1>{{ $t('migrate.title') }}</h1>
		<p>{{ $t('migrate.description') }}</p>
		<div class="migration-services-overview">
			<router-link
				v-for="{name, identifier} in availableMigrators"
				:key="identifier"
				:to="{name: 'migrate.service', params: {service: identifier}}"
			>
				<img :alt="name" :src="serviceIconSources[identifier]"/>
				{{ name }}
			</router-link>
		</div>
	</div>
</template>

<script>
import {getMigratorFromSlug, SERVICE_ICONS} from '../../helpers/migrator'


export default {
	name: 'migrate.service',
	mounted() {
		this.setTitle(this.$t('migrate.title'))
	},
	computed: {
		availableMigrators() {
			return this.$store.state.config.availableMigrators.map(getMigratorFromSlug)
		},
		serviceIconSources() {
			return this.availableMigrators.map(({identifier}) => SERVICE_ICONS[identifier]())
		},
	},
}
</script>

<style lang="scss" scoped>
.migration-services-overview {
  text-align: center;

  a {
    display: inline-block;
    width: 100px;
    text-transform: capitalize;
    margin-right: 1rem;

    img {
      display: block;
    }
  }
}
</style>