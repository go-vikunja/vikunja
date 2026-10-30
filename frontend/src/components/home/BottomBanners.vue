<template>
	<div
		ref="stack"
		class="bottom-banners"
	>
		<div class="bottom-banners__cards">
			<AddToHomeScreen />
			<UpdateNotification />
		</div>
		<div ref="bars">
			<DemoMode />
		</div>
	</div>
</template>

<script lang="ts" setup>
import {onScopeDispose, useTemplateRef, watchEffect} from 'vue'
import {useElementSize} from '@vueuse/core'

import AddToHomeScreen from '@/components/home/AddToHomeScreen.vue'
import UpdateNotification from '@/components/home/UpdateNotification.vue'
import DemoMode from '@/components/home/DemoMode.vue'
import {useIsMobile} from '@/composables/useIsMobile'

const stack = useTemplateRef<HTMLElement>('stack')
const bars = useTemplateRef<HTMLElement>('bars')
const {height} = useElementSize(stack, undefined, {box: 'border-box'})
const {height: barsHeight} = useElementSize(bars, undefined, {box: 'border-box'})
const isMobile = useIsMobile()

// Other bottom-fixed UI offsets itself by these to stay above the banners.
// On desktop the cards only occupy the start column, so the end edge just has the bars to clear.
const VARS = ['--bottom-banners-height', '--bottom-banners-end-height']
watchEffect(() => {
	const root = document.documentElement.style
	root.setProperty(VARS[0], `${height.value}px`)
	root.setProperty(VARS[1], `${isMobile.value ? height.value : barsHeight.value}px`)
})
onScopeDispose(() => VARS.forEach(v => document.documentElement.style.removeProperty(v)))
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
	// The stack spans the full width; only the banners themselves should catch clicks.
	pointer-events: none;

	> * > * {
		pointer-events: auto;
	}

	@media print {
		display: none;
	}
}

// Same column as the toasts, which stack on top.
.bottom-banners__cards {
	display: flex;
	flex-direction: column;
	align-items: flex-start;

	@media screen and (max-width: $tablet) {
		align-items: stretch;
	}
}
</style>
