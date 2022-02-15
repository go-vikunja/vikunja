<template>
	<div class="update-notification" v-if="updateAvailable">
		<p>{{ $t('update.available') }}</p>
		<x-button @click="refreshApp()" :shadow="false">
			{{ $t('update.do') }}
		</x-button>
	</div>
</template>

<script lang="ts">
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

<style lang="scss" scoped>
.update-notification {
	margin: 1rem;
	display: flex;
	align-items: center;
	background: $warning;
	padding: 0 0 0 .5rem;
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