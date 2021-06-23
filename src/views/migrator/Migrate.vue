<template>
	<div class="content">
		<h1>{{ $t('migrate.title') }}</h1>
		<p>{{ $t('migrate.description') }}</p>
		<div class="migration-services-overview">
			<router-link :key="m" :to="{name: 'migrate.service', params: {service: m}}" v-for="m in availableMigrators">
				<img :alt="m" :src="`/images/migration/${m}.png`"/>
				{{ m }}
			</router-link>
		</div>
	</div>
</template>

<script>
export default {
	name: 'migrate.service',
	mounted() {
		this.setTitle(this.$t('migrate.title'))
	},
	computed: {
		availableMigrators() {
			return this.$store.state.config.availableMigrators
		},
	},
}
</script>
