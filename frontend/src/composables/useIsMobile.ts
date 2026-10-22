import {createSharedComposable, useMediaQuery} from '@vueuse/core'

// Bulma's $tablet is 769px, so "mobile" is everything up to 768px.
const BULMA_MOBILE_BREAKPOINT = 768

export const useIsMobile = createSharedComposable(() => useMediaQuery(`(max-width: ${BULMA_MOBILE_BREAKPOINT}px)`))
