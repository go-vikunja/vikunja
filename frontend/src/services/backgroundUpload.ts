import AbstractService from './abstractService'
import ProjectModel from '@/models/project'

import type { IProject } from '@/modelTypes/IProject'
import type { IFile } from '@/modelTypes/IFile'

export default class BackgroundUploadService extends AbstractService {
	constructor() {
		super({
			create: '/projects/{projectId}/backgrounds/upload',
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
			this.getReplacedRoute(this.paths.create, {projectId}),
			file,
			'background',
		)
	}
}
