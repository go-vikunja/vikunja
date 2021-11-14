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
			<footer class="modal-card-foot is-flex is-justify-content-flex-end">
				<x-button
					type="secondary"
					@click.prevent.stop="$router.back()"
				>
					{{ $t('misc.close') }}
				</x-button>
			</footer>
		</card>
	</modal>

</template>

<script setup>
import {computed} from 'vue'

import { store } from '@/store'
import {VERSION as frontendVersion} from '@/version.json'

const apiVersion = computed(() => store.state.config.version)
</script>
