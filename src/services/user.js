import AbstractService from './abstractService'
import UserModel from '../models/user'
import moment from 'moment'

export default class UserService extends AbstractService {
	constructor() {
		super({
			getAll: '/users'
		})
	}

	processModel(model) {
		model.created = moment(model.created).toISOString()
		model.updated = moment(model.updated).toISOString()
		return model
	}

	modelFactory(data) {
		return new UserModel(data)
	}
}