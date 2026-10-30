<template>
	<div
		ref="stack"
		class="bottom-banners"
	>
		<AddToHomeScreen />
		<UpdateNotification />
		<DemoMode />
	</div>
</template>

<script lang="ts" setup>
import {onScopeDispose, useTemplateRef, watchEffect} from 'vue'
import {useElementSize} from '@vueuse/core'

import AddToHomeScreen from '@/components/home/AddToHomeScreen.vue'
import UpdateNotification from '@/components/home/UpdateNotification.vue'
import DemoMode from '@/components/home/DemoMode.vue'

const stack = useTemplateRef<HTMLElement>('stack')
const {height} = useElementSize(stack, undefined, {box: 'border-box'})

// Other bottom-fixed UI (toasts, floating buttons) offsets itself by this to stay above the banners.
watchEffect(() => document.documentElement.style.setProperty('--bottom-banners-height', `${height.value}px`))
onScopeDispose(() => document.documentElement.style.removeProperty('--bottom-banners-height'))
</script>

<style lang="scss" scoped>
.bottom-banners {
	position: fixed;
	inset-block-end: 0;
	inset-inline: 0;
	// FIXME: We should prevent usage of z-index or
	// at least define it centrally
	// the highest z-index of a modal is .hint-modal with 4500
	z-index: 5000;

	display: flex;
	flex-direction: column;
	align-items: center;
	// The stack spans the full width; only the banners themselves should catch clicks.
	pointer-events: none;

	> * {
		pointer-events: auto;
	}

	@media print {
		display: none;
	}
}
</style>
