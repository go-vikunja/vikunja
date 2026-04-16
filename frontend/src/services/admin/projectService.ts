import AbstractService from '@/services/abstractService'
import ProjectModel from '@/models/project'
import type {IProject} from '@/modelTypes/IProject'

export default class AdminProjectService extends AbstractService<IProject> {
	constructor() {
		super({
			getAll: '/admin/projects',
		})
	}

	modelFactory(data: Partial<IProject>) {
		return new ProjectModel(data)
	}

	async reassignOwner(projectId: IProject['id'], newOwnerId: IProject['owner']['id']) {
		const {data} = await this.http.patch(`/admin/projects/${projectId}/owner`, {owner_id: newOwnerId})
		return this.modelUpdateFactory(data)
	}
}
