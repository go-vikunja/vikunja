<template>
	<div class="update-notification" v-if="updateAvailable">
		<p>{{ $t('update.available') }}</p>
		<x-button @click="refreshApp()" :shadow="false">
			{{ $t('update.do') }}
		</x-button>
	</div>
</template>

<script>
export default {
	name: 'update',
	data() {
		return {
			updateAvailable: false,
			registration: null,
			refreshing: false,
		}
	},
	created() {
		document.addEventListener('swUpdated', this.showRefreshUI, {once: true})

		if (navigator && navigator.serviceWorker) {
			navigator.serviceWorker.addEventListener(
				'controllerchange', () => {
					if (this.refreshing) return
					this.refreshing = true
					window.location.reload()
				},
			)
		}
	},
	methods: {
		showRefreshUI(e) {
			console.log('recieved refresh event', e)
			this.registration = e.detail
			this.updateAvailable = true
		},
		refreshApp() {
			this.updateExists = false
			if (!this.registration || !this.registration.waiting) {
				return
			}
			// Notify the service worker to actually do the update
			this.registration.waiting.postMessage('skipWaiting')
		},
	},
}
</script>
