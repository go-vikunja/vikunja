import AbstractService from './abstractService'
import UserModel from '../models/user'
import {formatISO} from 'date-fns'

export default class ListUserService extends AbstractService {
	constructor() {
		super({
			getAll: '/lists/{listId}/listusers'
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