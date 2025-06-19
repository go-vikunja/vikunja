import {ref, toValue, watch, type MaybeRefOrGetter} from 'vue'
import ProjectService from '@/services/project'
import type {IProject} from '@/modelTypes/IProject'
import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

export function useProjectBackground(project: MaybeRefOrGetter<IProject | null>) {
	const background = ref<string | null>(null)
	const backgroundLoading = ref(false)
	const blurHashUrl = ref('')

	watch(
		() => [
			toValue(project)?.id ?? null,
			toValue(project)?.backgroundBlurHash ?? null,
		] as [IProject['id'] | null, IProject['backgroundBlurHash'] | null],
		async ([projectId, blurHash], oldValue) => {
			const projectValue = toValue(project)
			if (
				projectValue === null ||
				!projectValue.backgroundInformation ||
				backgroundLoading.value
			) {
				return
			}

			const [oldProjectId, oldBlurHash] = oldValue || []
			if (
				oldValue !== undefined &&
				projectId === oldProjectId && blurHash === oldBlurHash
			) {
				// project hasn't changed
				return
			}

			backgroundLoading.value = true

			try {
				const blurHashPromise = getBlobFromBlurHash(blurHash).then((blurHash) => {
					blurHashUrl.value = blurHash ? window.URL.createObjectURL(blurHash) : ''
				})

				const projectService = new ProjectService()
				const backgroundPromise = projectService.background(projectValue).then((result) => {
					background.value = result
				})
				await Promise.all([blurHashPromise, backgroundPromise])
			} finally {
				backgroundLoading.value = false
			}
		},
		{immediate: true},
	)

	return {
		background,
		blurHashUrl,
		backgroundLoading,
	}
}
