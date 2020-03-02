import AbstractService from './abstractService'
import UserModel from '../models/user'
import {formatISO} from 'date-fns'

export default class ListUserService extends AbstractService {
	constructor() {
		super({
			getAll: '/lists/{listID}/listusers'
		})
	}

	processModel(model) {
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)
		return model
	}

	modelFactory(data) {
		return new UserModel(data)
	}
}