<template>
	<div
		v-if="shouldShowMessage"
		class="add-to-home-screen"
		:class="{'has-update-available': hasUpdateAvailable}"
	>
		<Icon
			icon="arrow-up-from-bracket"
			class="add-icon"
		/>
		<p>
			{{ $t('home.addToHomeScreen') }}
		</p>
		<BaseButton
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
import {useBaseStore} from '@/stores/base'

const baseStore = useBaseStore()

const hideMessage = useLocalStorage('hideAddToHomeScreenMessage', false)
const hasUpdateAvailable = computed(() => baseStore.updateAvailable)

const shouldShowMessage = computed(() => {
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
	padding: .5rem 1rem;
	background: var(--grey-900);
	border-radius: $radius;
	font-size: .9rem;
	color: var(--grey-200);

	@media screen and (min-width: $tablet) {
		display: none;
	}

	@media print {
		display: none;
	}
	
	&.has-update-available {
		inset-block-end: 5rem;
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
