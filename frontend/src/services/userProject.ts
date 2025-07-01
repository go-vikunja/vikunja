import AbstractService from './abstractService'
import UserProjectModel from '@/models/userProject'
import type {IUserProject} from '@/modelTypes/IUserProject'
import type {IUser} from '@/modelTypes/IUser'
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

	modelFactory(data: Partial<IUserProject>): IUserProject {
		return new UserProjectModel(data) as IUserProject
	}

	modelGetAllFactory(data: Partial<IUser>): IUser {
		return new UserModel(data)
	}
}
