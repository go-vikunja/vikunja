<template>
	<div
		v-if="updateAvailable"
		class="update-notification"
	>
		<p class="update-notification__message">
			{{ $t('update.available') }}
		</p>
		<XButton
			:shadow="false"
			:wrap="false"
			@click="refreshApp()"
		>
			{{ $t('update.do') }}
		</XButton>
	</div>
</template>

<script lang="ts" setup>
import {computed, ref} from 'vue'
import {useBaseStore} from '@/stores/base'

const baseStore = useBaseStore()

const updateAvailable = computed(() => baseStore.updateAvailable)
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
	baseStore.setUpdateAvailable(true)
}

function refreshApp() {
	baseStore.setUpdateAvailable(false)
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
	inset-block-end: 1rem;
	inset-inline: 1rem;
	max-inline-size: max-content;
	margin-inline: auto;

	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
	padding: .5rem .5rem .5rem 1rem;
	background: $warning;
	border-radius: $radius;
	font-size: .9rem;
	color: hsl(220.9, 39.3%, 11%); // color copied to avoid it changing in dark mode
}

.update-notification__message {
	inline-size: 100%;
	text-align: center;
}
</style>
