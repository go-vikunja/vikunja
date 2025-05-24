import {computed, toValue} from 'vue'

import {useTitle as useTitleVueUse, type UseTitleOptions, type ReadonlyRefOrGetter, type MaybeRef, type MaybeRefOrGetter} from '@vueuse/core'

export function useTitle(
	newTitle:
		| ReadonlyRefOrGetter<string | null | undefined>
		| MaybeRef<string | null | undefined>
		| MaybeRefOrGetter<string | null | undefined> = null,
	options?: UseTitleOptions,
) {
	const pageTitle = computed(() => toValue(newTitle))

	const completeTitle = computed(() =>
		(typeof pageTitle.value === 'undefined' || pageTitle.value === '')
			? 'Vikunja'
			: `${pageTitle.value} | Vikunja`,
	)

	return useTitleVueUse(completeTitle, options)
}
