import type {MaybeRefOrGetter} from 'vue'
import ProjectService from '@/services/project'
import type {IProject} from '@/modelTypes/IProject'
import {useProjectImage} from '@/composables/useProjectBackground'

export function useProjectCardBackground(project: MaybeRefOrGetter<IProject | null>) {
	const {image, blurHashUrl, imageLoading} = useProjectImage(
		project,
		p => p.cardBackgroundInformation,
		p => p.cardBackgroundBlurHash,
		p => new ProjectService().cardBackground(p),
	)

	return {
		background: image,
		blurHashUrl,
		backgroundLoading: imageLoading,
	}
}
