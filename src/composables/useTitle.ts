import { computed } from 'vue'
import type { Ref } from 'vue'

import {useTitle as useTitleVueUse, resolveRef} from '@vueuse/core'

type UseTitleParameters = Parameters<typeof useTitleVueUse>

export function useTitle(...args: UseTitleParameters) {

	const [newTitle, ...restArgs] = args

  const pageTitle = resolveRef(newTitle) as Ref<string>

	const completeTitle = computed(() => 
		(typeof pageTitle.value === 'undefined' || pageTitle.value === '')
		? 'Vikunja'
		: `${pageTitle.value} | Vikunja`,
	)

	return useTitleVueUse(completeTitle, ...restArgs)
}