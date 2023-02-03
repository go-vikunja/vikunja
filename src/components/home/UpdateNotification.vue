<template>
	<div class="update-notification" v-if="updateAvailable">
		<p class="update-notification__message">{{ $t('update.available') }}</p>
		<x-button
			@click="refreshApp()"
			:shadow="false"
			:wrap="false"
			>
			{{ $t('update.do') }}
		</x-button>
	</div>
</template>

<script lang="ts" setup>
import {ref} from 'vue'

const updateAvailable = ref(false)
const registration = ref(null)
const refreshing = ref(false)

document.addEventListener('swUpdated', showRefreshUI, {once: true})

navigator?.serviceWorker?.addEventListener(
	'controllerchange', () => {
		if (refreshing.value) return
		refreshing.value = true
		window.location.reload()
	},
)

function showRefreshUI(e: Event) {
	console.log('recieved refresh event', e)
	registration.value = e.detail
	updateAvailable.value = true
}

function refreshApp() {
	updateAvailable.value = false
	if (!registration.value || !registration.value.waiting) {
		return
	}
	// Notify the service worker to actually do the update
	registration.value.waiting.postMessage('skipWaiting')
}
</script>

<style lang="scss" scoped>
.update-notification {
	position: fixed;
	// FIXME: We should prevent usage of z-index or
	// at least define it centrally
	// the highest z-index of a modal is .hint-modal with 4500
	z-index: 5000;
	bottom: 1rem;
	inset-inline: 1rem;
	max-width: max-content;
	margin-inline: auto;

	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
	padding: .5rem;
	background: $warning;
	border-radius: $radius;
	font-size: .9rem;
	color: var(--grey-900);

}

.update-notification__message {
	width: 100%;
	text-align: center;
}
</style>