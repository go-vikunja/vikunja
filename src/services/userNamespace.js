import AbstractService from './abstractService'
import UserNamespaceModel from '../models/userNamespace'
import UserModel from '../models/user'
import {formatISO} from 'date-fns'

export default class UserNamespaceService extends AbstractService {
	constructor() {
		super({
			create: '/namespaces/{namespaceID}/users',
			getAll: '/namespaces/{namespaceID}/users',
			update: '/namespaces/{namespaceID}/users/{userId}',
			delete: '/namespaces/{namespaceID}/users/{userId}',
		})
	}

	processModel(model) {
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)
		return model
	}

	modelFactory(data) {
		return new UserNamespaceModel(data)
	}

	modelGetAllFactory(data) {
		return new UserModel(data)
	}
}