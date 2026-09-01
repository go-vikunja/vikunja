import {
	computed,
	onBeforeUnmount,
	ref,
	toValue,
	watch,
	type MaybeRefOrGetter,
} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import type {ProjectResponse} from '@/client/queries/projects'
import {projectBackgroundQuery} from '@/client/queries/projectBackgrounds'
import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

export function useProjectBackground(project: MaybeRefOrGetter<ProjectResponse | null>) {
	const projectId = computed(() => toValue(project)?.id ?? 0)
	const hasBackground = computed(() => Boolean(toValue(project)?.background_information))
	const blurHash = computed(() => toValue(project)?.background_blur_hash ?? '')
	const query = useQuery(computed(() => ({
		...projectBackgroundQuery(projectId.value),
		enabled: hasBackground.value,
	})))

	const background = ref<string | null>(null)
	const blurHashUrl = ref('')

	function clearBackground() {
		if (background.value !== null) {
			window.URL.revokeObjectURL(background.value)
			background.value = null
		}
	}

	function clearBlurHash() {
		if (blurHashUrl.value !== '') {
			window.URL.revokeObjectURL(blurHashUrl.value)
			blurHashUrl.value = ''
		}
	}

	watch([projectId, hasBackground], () => {
		clearBackground()
		if (!hasBackground.value) {
			clearBlurHash()
		}
	}, {immediate: true})

	watch(query.data, blob => {
		clearBackground()
		if (hasBackground.value && blob) {
			background.value = window.URL.createObjectURL(blob)
		}
	}, {immediate: true})

	watch(
		[projectId, hasBackground, blurHash],
		async ([, projectHasBackground, projectBlurHash], _, onCleanup) => {
			let active = true
			onCleanup(() => {
				active = false
			})

			clearBlurHash()
			if (!projectHasBackground || projectBlurHash === '') {
				return
			}

			try {
				const blob = await getBlobFromBlurHash(projectBlurHash)
				if (active && blob) {
					blurHashUrl.value = window.URL.createObjectURL(blob)
				}
			} catch (e) {
				console.error('Error generating blur hash preview', e)
			}
		},
		{immediate: true},
	)

	onBeforeUnmount(() => {
		clearBackground()
		clearBlurHash()
	})

	return {
		background,
		blurHashUrl,
	}
}
