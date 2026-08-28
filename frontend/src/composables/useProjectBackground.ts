import {ref, toValue, watch, type MaybeRefOrGetter} from 'vue'
import ProjectService from '@/services/project'
import type {IProject} from '@/modelTypes/IProject'
import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

// Shared by useProjectBackground and useProjectCardBackground — same watch/blurhash
// logic, parameterized by which pair of project fields to watch and how to fetch
// the image itself.
export function useProjectImage(
	project: MaybeRefOrGetter<IProject | null>,
	getInformation: (project: IProject) => unknown,
	getBlurHash: (project: IProject) => string,
	fetchImage: (project: IProject) => Promise<string>,
) {
	const image = ref<string | null>(null)
	const imageLoading = ref(false)
	const blurHashUrl = ref('')

	watch(
		() => [
			toValue(project)?.id ?? null,
			toValue(project) ? getBlurHash(toValue(project) as IProject) : null,
		] as [IProject['id'] | null, string | null],
		async ([projectId, blurHash], oldValue) => {
			const projectValue = toValue(project)
			if (
				projectValue === null ||
				!getInformation(projectValue) ||
				imageLoading.value
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

			imageLoading.value = true

			try {
				const blurHashPromise = getBlobFromBlurHash(blurHash).then((blurHash) => {
					blurHashUrl.value = blurHash ? window.URL.createObjectURL(blurHash) : ''
				})

				const imagePromise = fetchImage(projectValue).then((result) => {
					image.value = result
				})
				await Promise.all([blurHashPromise, imagePromise])
			} finally {
				imageLoading.value = false
			}
		},
		{immediate: true},
	)

	return {
		image,
		blurHashUrl,
		imageLoading,
	}
}

export function useProjectBackground(project: MaybeRefOrGetter<IProject | null>) {
	const {image, blurHashUrl, imageLoading} = useProjectImage(
		project,
		p => p.backgroundInformation,
		p => p.backgroundBlurHash,
		p => new ProjectService().background(p),
	)

	return {
		background: image,
		blurHashUrl,
		backgroundLoading: imageLoading,
	}
}
