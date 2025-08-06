<template>
	<div class="content">
		<h1>{{ $t('migrate.title') }}</h1>
		<p>{{ $t('migrate.description') }}</p>
		<div class="migration-services">
			<RouterLink
				v-for="{name, id, icon} in availableMigrators"
				:key="id"
				class="migration-service-link"
				:to="{name: 'migrate.service', params: {service: id}}"
			>
				<img
					class="migration-service-image"
					:alt="name"
					:src="icon"
				>
				{{ name }}
			</RouterLink>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import {MIGRATORS} from './migrators'
import {useTitle} from '@/composables/useTitle'
import {useConfigStore} from '@/stores/config'

const {t} = useI18n({useScope: 'global'})

useTitle(() => t('migrate.title'))

const configStore = useConfigStore()
const availableMigrators = computed(() => configStore.availableMigrators
	.map((id) => MIGRATORS[id])
	.filter((item) => Boolean(item)),
)
</script>

<style lang="scss" scoped>
.migration-services {
  text-align: center;
}

.migration-service-link {
    display: inline-block;
    inline-size: 100px;
    text-transform: capitalize;
    margin-inline-end: 1rem;
}

.migration-service-image {
	display: block;
}
</style>
