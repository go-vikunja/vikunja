import {
	computed,
	onBeforeUnmount,
	ref,
	toValue,
	watch,
	type MaybeRefOrGetter,
} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import type {Project} from '@/client/generated'
import {projectBackgroundQuery} from '@/client/queries/projectBackgrounds'
import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

type ProjectWithBackground = Pick<
	Project,
	'id' | 'background_information' | 'background_blur_hash'
>

export function useProjectBackground(project: MaybeRefOrGetter<ProjectWithBackground | null>) {
	const projectId = computed(() => toValue(project)?.id ?? 0)
	const hasBackground = computed(() => Boolean(toValue(project)?.background_information))
	const blurHash = computed(() => toValue(project)?.background_blur_hash ?? '')
	const query = useQuery(computed(() => ({
		...projectBackgroundQuery(projectId.value),
		enabled: projectId.value > 0 && hasBackground.value,
	})))

	const background = ref<string | null>(null)
	const blurHashUrl = ref('')
	const blurHashLoading = ref(false)

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
				blurHashLoading.value = false
				return
			}

			blurHashLoading.value = true
			try {
				const blob = await getBlobFromBlurHash(projectBlurHash)
				if (active && blob) {
					blurHashUrl.value = window.URL.createObjectURL(blob)
				}
			} finally {
				if (active) {
					blurHashLoading.value = false
				}
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
		backgroundLoading: computed(() =>
			hasBackground.value && (query.isPending.value || blurHashLoading.value),
		),
	}
}
