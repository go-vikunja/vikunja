import {computed, shallowRef, toValue, type MaybeRefOrGetter} from 'vue'
import {useWindowSize} from '@vueuse/core'

// Viewport rects of the visible banners in BottomBanners.vue.
export const bottomBannerRects = shallowRef<DOMRect[]>([])

export interface HorizontalRange {
	left: number
	right: number
}

/**
 * How far from the viewport bottom something spanning `range` must sit to clear
 * every banner below it. Banners elsewhere on the x-axis don't count.
 */
export function useBottomBannersOffset(range: MaybeRefOrGetter<HorizontalRange>) {
	const {height: viewportHeight} = useWindowSize()

	return computed(() => {
		const {left, right} = toValue(range)
		const covering = bottomBannerRects.value.filter(r => r.left < right && r.right > left)
		return Math.max(0, ...covering.map(r => viewportHeight.value - r.top))
	})
}
