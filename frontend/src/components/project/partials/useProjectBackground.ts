import {ref, watch, type ShallowReactive} from 'vue'
import ProjectService from '@/services/project'
import type {IProject} from '@/modelTypes/IProject'
import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

export function useProjectBackground(project: ShallowReactive<IProject>) {
	const background = ref<string | null>(null)
	const backgroundLoading = ref(false)
	const blurHashUrl = ref('')

	watch(
		() => [project.id, project.backgroundBlurHash] as [IProject['id'], IProject['backgroundBlurHash']],
		async ([projectId, blurHash], oldValue) => {
			if (
				project === null ||
				!project.backgroundInformation ||
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
				const backgroundPromise = projectService.background(project).then((result) => {
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
