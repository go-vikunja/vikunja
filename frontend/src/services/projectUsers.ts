import AbstractService from './abstractService'
import UserModel from '../models/user'

export default class ProjectUserService extends AbstractService {
	constructor() {
		super({
			getAll: '/projects/{projectId}/projectusers',
		})
	}

	modelFactory(data) {
		return new UserModel(data)
	}
}
