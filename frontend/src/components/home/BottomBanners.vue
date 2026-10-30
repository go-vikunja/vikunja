<template>
	<div
		ref="stack"
		class="bottom-banners"
	>
		<AddToHomeScreen />
		<UpdateNotification />
		<DemoMode class="bottom-banners__bar" />
	</div>
</template>

<script lang="ts" setup>
import {onScopeDispose, useTemplateRef, watch, watchEffect} from 'vue'
import {useMutationObserver, useResizeObserver, useWindowSize} from '@vueuse/core'

import AddToHomeScreen from '@/components/home/AddToHomeScreen.vue'
import UpdateNotification from '@/components/home/UpdateNotification.vue'
import DemoMode from '@/components/home/DemoMode.vue'
import {bottomBannerRects, useBottomBannersOffset} from '@/composables/useBottomBanners'
import {useIsRtl} from '@/composables/useIsRtl'

const stack = useTemplateRef<HTMLElement>('stack')
const {width: viewportWidth, height: viewportHeight} = useWindowSize()

function measure() {
	bottomBannerRects.value = [...(stack.value?.children ?? [])]
		.map(el => el.getBoundingClientRect())
		.filter(r => r.width > 0 && r.height > 0)
}

useResizeObserver(stack, measure)
useMutationObserver(stack, measure, {childList: true, subtree: true})
watch([viewportWidth, viewportHeight], measure)
onScopeDispose(() => bottomBannerRects.value = [])

// For floating controls pinned to the end edge (keyboard shortcuts, scroll buttons).
const EDGE_ZONE = 64
const isRtl = useIsRtl()
const endOffset = useBottomBannersOffset(() => isRtl.value
	? {left: 0, right: EDGE_ZONE}
	: {left: viewportWidth.value - EDGE_ZONE, right: viewportWidth.value})

watchEffect(() => document.documentElement.style.setProperty('--bottom-banners-end-height', `${endOffset.value}px`))
onScopeDispose(() => document.documentElement.style.removeProperty('--bottom-banners-end-height'))
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

	@media screen and (max-width: $tablet) {
		align-items: stretch;
	}

	@media print {
		display: none;
	}
}

.bottom-banners__bar {
	align-self: stretch;
}
</style>
