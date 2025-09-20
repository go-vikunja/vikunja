import AbstractService from './abstractService'
import UserModel from '../models/user'
import type {IUser} from '@/modelTypes/IUser'

export default class ProjectUserService extends AbstractService<IUser> {
	constructor() {
		super({
			getAll: '/projects/{projectId}/projectusers',
		})
	}

	modelFactory(data: Partial<IUser>) {
		return new UserModel(data)
	}
}
