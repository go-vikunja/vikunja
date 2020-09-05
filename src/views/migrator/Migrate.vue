<template>
	<div class="content">
		<h1>Import your data from other services to Vikunja</h1>
		<p>Click on the logo of one of the third-party services below to get started.</p>
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
		this.setTitle('Import your data to Vikunja')
	},
	computed: {
		availableMigrators() {
			return this.$store.state.config.availableMigrators
		},
	},
}
</script>
