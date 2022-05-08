<template>
	<div class="update-notification" v-if="updateAvailable">
		<p>{{ $t('update.available') }}</p>
		<x-button @click="refreshApp()" :shadow="false">
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

if (navigator && navigator.serviceWorker) {
	navigator.serviceWorker.addEventListener(
		'controllerchange', () => {
			if (refreshing.value) return
			refreshing.value = true
			window.location.reload()
		},
	)
}

function showRefreshUI(e) {
	console.log('recieved refresh event', e)
	registration.value = e.detail
	updateAvailable.value = true
}

function refreshApp() {
	if (!registration.value || !registration.value.waiting) {
		return
	}
	// Notify the service worker to actually do the update
	registration.value.waiting.postMessage('skipWaiting')
}
</script>

<style lang="scss" scoped>
.update-notification {
	margin: 1rem;
	display: flex;
	align-items: center;
	background: $warning;
	padding: .25rem .5rem;
	border-radius: $radius;
	font-size: .9rem;
	color: var(--grey-900);
	justify-content: space-between;

	@media screen and (max-width: $desktop) {
		position: fixed;
		bottom: 1rem;
		margin: 0;
		width: 450px;
		left: calc(50vw - 225px);
	}

	@media screen and (max-width: $tablet) {
		position: fixed;
		left: 1rem;
		right: 1rem;
		bottom: 1rem;
		width: auto;
	}

	p {
		text-align: center;
		width: 100%;
	}

	> * + * {
		margin-left: .5rem;
	}
}

.dark .update-notification {
	color: var(--grey-200);
}
</style>