import AbstractService from '@/services/abstractService'
import type {IAbstract} from '@/modelTypes/IAbstract'
import ProjectViewModel from '@/models/projectView'
import type {IProjectView} from '@/modelTypes/IProjectView'

export default class ProjectViewService extends AbstractService<IProjectView> {
	constructor() {
		super({
			get: '/projects/{projectId}/views/{id}',
			getAll: '/projects/{projectId}/views',
			create: '/projects/{projectId}/views',
			update: '/projects/{projectId}/views/{id}',
			delete: '/projects/{projectId}/views/{id}',
		})
	}

	modelFactory(data: Partial<IAbstract>): ProjectViewModel {
		return new ProjectViewModel(data)
	}
}
