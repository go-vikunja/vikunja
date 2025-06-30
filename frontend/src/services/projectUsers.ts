import AbstractService from './abstractService'
import UserModel from '../models/user'

export default class ProjectUserService extends AbstractService {
	constructor() {
		super({
			getAll: '/projects/{projectId}/projectusers',
		})
	}

	modelFactory(data: any) {
		return new UserModel(data)
	}
}
