<template>
	<div
		v-if="shouldShowMessage"
		class="add-to-home-screen"
	>
		<Icon
			icon="arrow-up-from-bracket"
			class="add-icon"
		/>
		<p>
			{{ $t('home.addToHomeScreen') }}
		</p>
		<BaseButton
			:aria-label="$t('misc.closeBanner')"
			class="hide-button"
			@click="() => hideMessage = true"
		>
			<Icon icon="x" />
		</BaseButton>
	</div>
</template>

<script lang="ts" setup>
import BaseButton from '@/components/base/BaseButton.vue'
import {useLocalStorage} from '@vueuse/core'
import {computed} from 'vue'
import {useAuthStore} from '@/stores/auth'

const authStore = useAuthStore()

const hideMessage = useLocalStorage('hideAddToHomeScreenMessage', false)

const shouldShowMessage = computed(() => {
	if (!authStore.authenticated) {
		return false
	}

	if (hideMessage.value) {
		return false
	}

	if (typeof window !== 'undefined' && window.matchMedia('(display-mode: standalone)').matches) {
		return false
	}

	return true
})
</script>

<style lang="scss" scoped>
.add-to-home-screen {
	max-inline-size: calc(100% - 1.5rem);
	margin: 0 .75rem .75rem;

	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
	padding: .5rem 1rem;
	background: var(--grey-900);
	border-radius: $radius;
	font-size: .9rem;
	color: var(--grey-200);

	@media screen and (min-width: $tablet) {
		display: none;
	}
}

.add-icon {
	color: var(--primary-light);
}

.hide-button {
	padding: .25rem .5rem;
	cursor: pointer;
}
</style>
