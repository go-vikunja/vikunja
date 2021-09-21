<template>
	<transition name="fade">
		<div class="modal-mask hint-modal">
			<div
				class="modal-container"
				@click.self="$router.back()"
				@shortkey="$router.back()"
				v-shortkey="['esc']"
			>
				<div class="modal-content">
					<card
						class="has-background-white has-no-shadow"
						:title="$t('about.title')"
						:has-close="true"
						close-icon="times"
						@close="$router.back()"
						:padding="false"
					>
						<div class="p-4">
							<p>
								{{ $t('about.frontendVersion', {version: this.frontendVersion}) }}
							</p>
							<p>
								{{ $t('about.apiVersion', {version: this.apiVersion}) }}
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
				</div>
			</div>
		</div>
	</transition>
</template>

<script>
import {VERSION} from '../version.json'

export default {
	name: 'About',
	computed: {
		frontendVersion() {
			return VERSION
		},
		apiVersion() {
			return this.$store.state.config.version
		},
	},
}
</script>
