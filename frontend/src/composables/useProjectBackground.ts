import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useObjectUrl} from '@vueuse/core'

import type {ProjectResponse} from '@/client/queries/projects'
import {projectBackgroundQuery} from '@/client/queries/projectBackgrounds'
import {useBlurHashUrl} from '@/composables/useBlurHashUrl'

export function useProjectBackground(project: MaybeRefOrGetter<ProjectResponse | null>) {
	const projectId = computed(() => toValue(project)?.id ?? 0)
	const hasBackground = computed(() => Boolean(toValue(project)?.background_information))
	const blurHash = computed(() => toValue(project)?.background_blur_hash ?? '')
	const query = useQuery(computed(() => ({
		...projectBackgroundQuery(projectId.value),
		enabled: hasBackground.value,
	})))

	// Disabled queries keep serving cached data, so gate on hasBackground.
	const background = useObjectUrl(computed(() => hasBackground.value ? query.data.value : undefined))
	const blurHashUrl = useBlurHashUrl(computed(() => hasBackground.value ? blurHash.value : ''))

	return {
		background,
		blurHashUrl,
	}
}
