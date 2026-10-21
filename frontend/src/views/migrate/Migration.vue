<template>
	<div class="content">
		<h1>{{ $t('migrate.title') }}</h1>
		<p>{{ $t('migrate.description') }}</p>
		<div
			ref="servicesContainer"
			class="migration-services"
		>
			<div
				v-for="(row, index) in migratorRows"
				:key="index"
				class="migration-services-row"
			>
				<RouterLink
					v-for="{name, id, icon, kind} in row"
					:key="id"
					class="migration-service-link"
					:to="kind === 'csv' ? {name: 'migrate.csv'} : {name: 'migrate.service', params: {service: id}}"
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
	</div>
</template>

<script setup lang="ts">
import {computed, ref, useTemplateRef} from 'vue'
import {useI18n} from 'vue-i18n'
import {useResizeObserver} from '@vueuse/core'

import {MIGRATORS, type Migrator} from './migrators'
import {useTitle} from '@/composables/useTitle'
import {useConfigStore} from '@/stores/config'
import {balancedRows} from '@/helpers/balancedRows'

const {t} = useI18n({useScope: 'global'})

useTitle(() => t('migrate.title'))

const configStore = useConfigStore()
const availableMigrators = computed(() => configStore.available_migrators
	.filter((id): id is keyof typeof MIGRATORS => Object.prototype.hasOwnProperty.call(MIGRATORS, id))
	.map((id): Migrator => MIGRATORS[id]),
)

// Keep in sync with the link size and row gap in the styles below.
const LINK_WIDTH_PX = 100
const GAP_PX = 16

const servicesContainer = useTemplateRef('servicesContainer')
const containerWidth = ref(0)
useResizeObserver(servicesContainer, ([entry]) => {
	// Re-rendering rows within the observer callback throws "ResizeObserver loop completed"
	requestAnimationFrame(() => containerWidth.value = entry.contentRect.width)
})
const migratorRows = computed(() => {
	// Not measured yet on first render
	const maxPerRow = containerWidth.value === 0
		? availableMigrators.value.length
		: Math.floor((containerWidth.value + GAP_PX) / (LINK_WIDTH_PX + GAP_PX))
	return balancedRows(availableMigrators.value, maxPerRow)
})
</script>

<style lang="scss" scoped>
.migration-services {
	display: flex;
	flex-direction: column;
	gap: 1rem;
	text-align: center;
}

.migration-services-row {
	display: flex;
	justify-content: center;
	gap: 1rem;
}

.migration-service-link {
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-end;
    inline-size: 100px;
    text-transform: capitalize;
}

.migration-service-image {
	display: block;
	max-block-size: 80px;
	inline-size: auto;
	margin-block-end: 0.5rem;
}
</style>
