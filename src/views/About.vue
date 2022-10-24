<template>
	<modal
		@close="$router.back()"
		transition-name="fade"
		variant="hint-modal"
	>
		<card
			class="has-no-shadow"
			:title="$t('about.title')"
			:has-close="true"
			@close="$router.back()"
			:padding="false"
		>
			<div class="p-4">
				<p>
					{{ $t('about.frontendVersion', {version: frontendVersion}) }}
				</p>
				<p>
					{{ $t('about.apiVersion', {version: apiVersion}) }}
				</p>
			</div>
			<template #footer>
				<x-button
					variant="secondary"
					@click.prevent.stop="$router.back()"
				>
					{{ $t('misc.close') }}
				</x-button>
			</template>
		</card>
	</modal>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import {VERSION} from '@/version.json'
import {useConfigStore} from '@/stores/config'

const configStore = useConfigStore()
const apiVersion = computed(() => configStore.version)
const frontendVersion = VERSION
</script>
