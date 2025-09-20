import AbstractService from './abstractService'
import UserProjectModel from '@/models/userProject'
import type {IUserProject} from '@/modelTypes/IUserProject'

export default class UserProjectService extends AbstractService<IUserProject> {
	constructor() {
		super({
			create: '/projects/{projectId}/users',
			getAll: '/projects/{projectId}/users',
			update: '/projects/{projectId}/users/{username}',
			delete: '/projects/{projectId}/users/{username}',
		})
	}

	modelFactory(data: Partial<IUserProject>) {
		return new UserProjectModel(data)
	}

	modelGetAllFactory(data: Partial<IUserProject>) {
		return new UserProjectModel(data)
	}
}
