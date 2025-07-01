import AbstractService from './abstractService'
import UserModel from '../models/user'
import type {IUser} from '@/modelTypes/IUser'

export default class ProjectUserService extends AbstractService {
	constructor() {
		super({
			getAll: '/projects/{projectId}/projectusers',
		})
	}

	modelFactory(data: Partial<IUser>): IUser {
		return new UserModel(data)
	}
}
