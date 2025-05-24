import AbstractService from './abstractService'
import UserProjectModel from '@/models/userProject'
import type {IUserProject} from '@/modelTypes/IUserProject'
import UserModel from '@/models/user'

export default class UserProjectService extends AbstractService<IUserProject> {
	constructor() {
		super({
			create: '/projects/{projectId}/users',
			getAll: '/projects/{projectId}/users',
			update: '/projects/{projectId}/users/{userId}',
			delete: '/projects/{projectId}/users/{userId}',
		})
	}

	modelFactory(data) {
		return new UserProjectModel(data)
	}

	modelGetAllFactory(data) {
		return new UserModel(data)
	}
}
