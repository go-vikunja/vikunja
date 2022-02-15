<template>
	<div class="content">
		<h1>{{ $t('migrate.title') }}</h1>
		<p>{{ $t('migrate.description') }}</p>
		<div class="migration-services">
			<router-link
				v-for="{name, id, icon} in availableMigrators"
				:key="id"
				class="migration-service-link"
				:to="{name: 'migrate.service', params: {service: id}}"
			>
				<img
					class="migration-service-image"
					:alt="name"
					:src="icon"
				/>
				{{ name }}
			</router-link>
		</div>
	</div>
</template>

<script lang="ts">
import {MIGRATORS} from './migrators'

export default {
	name: 'Migrate',
	mounted() {
		this.setTitle(this.$t('migrate.title'))
	},
	computed: {
		availableMigrators() {
			return this.$store.state.config.availableMigrators
				.map((id) => MIGRATORS[id])
				.filter((item) => Boolean(item))
		},
	},
}
</script>

<style lang="scss" scoped>
.migration-services {
  text-align: center;
}

.migration-service-link {
    display: inline-block;
    width: 100px;
    text-transform: capitalize;
    margin-right: 1rem;

}

.migration-service-image {
	display: block;
}
</style>