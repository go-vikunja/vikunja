import AbstractService from './abstractService'
import ProjectModel from '@/models/project'

import type { IProject } from '@/modelTypes/IProject'
import type { IFile } from '@/modelTypes/IFile'
import type { IAbstract } from '@/modelTypes/IAbstract'

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
	create(model: IAbstract): Promise<IAbstract>
	create(projectId: IProject['id'], file: IFile | File): Promise<IProject>
	create(modelOrProjectId: IAbstract | IProject['id'], file?: IFile | File): Promise<IAbstract | IProject> {
		if (typeof modelOrProjectId === 'object') {
			return super.create(modelOrProjectId)
		}
		
		return this.uploadFile(
			this.getReplacedRoute(this.paths.create, {projectId: modelOrProjectId}),
			file as File,
			'background',
		)
	}
}
