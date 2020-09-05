import AbstractService from './abstractService'
import UserModel from '../models/user'
import {formatISO} from 'date-fns'

export default class UserService extends AbstractService {
	constructor() {
		super({
			getAll: '/users',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new UserModel(data)
	}
}