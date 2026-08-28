import AbstractService from './abstractService'
import ProjectModel from '@/models/project'
import {apiV2Url} from '@/helpers/fetcher'

import type { IProject } from '@/modelTypes/IProject'
import type { IFile } from '@/modelTypes/IFile'

// The card image upload endpoint only exists on /api/v2, hence the absolute URL.
export default class CardBackgroundUploadService extends AbstractService {
	constructor() {
		super({
			create: 'projects/{projectId}/card-backgrounds/upload',
		})
	}

	useCreateInterceptor() {
		return false
	}

	modelCreateFactory(data: Partial<IProject>) {
		return new ProjectModel(data)
	}

	/**
	 * Uploads a file to the server
	 */
	create(projectId: IProject['id'], file: IFile) {
		return this.uploadFile(
			apiV2Url(this.getReplacedRoute(this.paths.create, {projectId})),
			file,
			'background',
		)
	}
}
