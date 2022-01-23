import {computed, watch, readonly} from 'vue'
import {useStorage, createSharedComposable, BasicColorSchema, usePreferredColorScheme, tryOnMounted} from '@vueuse/core'

const STORAGE_KEY = 'color-scheme'

const DEFAULT_COLOR_SCHEME_SETTING: BasicColorSchema = 'light'

const CLASS_DARK = 'dark'
const CLASS_LIGHT = 'light'

// This is built upon the vueuse useDark
// Main differences:
// - usePreferredColorScheme
// - doesn't allow setting via the `isDark` ref.
// - instead the store is exposed
// - value is synced via `createSharedComposable`
// https://github.com/vueuse/vueuse/blob/main/packages/core/useDark/index.ts 
export const useColorScheme = createSharedComposable(() => {
	const store = useStorage<BasicColorSchema>(STORAGE_KEY, DEFAULT_COLOR_SCHEME_SETTING)

	const preferredColorScheme = usePreferredColorScheme()

	const isDark = computed<boolean>(() => {
		if (store.value !== 'auto') {
			return store.value === 'dark'
		}

		const autoColorScheme = preferredColorScheme.value === 'no-preference' 
			? DEFAULT_COLOR_SCHEME_SETTING
			: preferredColorScheme.value
		return autoColorScheme === 'dark'
	})

    function onChanged(v: boolean) {
		const el = window?.document.querySelector('html')
		el?.classList.toggle(CLASS_DARK, v)
		el?.classList.toggle(CLASS_LIGHT, !v)
	}

    watch(isDark, onChanged, { flush: 'post' })

	tryOnMounted(() => onChanged(isDark.value))

	return {
		store,
		isDark: readonly(isDark),
	}
})