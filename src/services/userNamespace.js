import AbstractService from './abstractService'
import UserNamespaceModel from '../models/userNamespace'
import UserModel from '../models/user'
import {formatISO} from 'date-fns'

export default class UserNamespaceService extends AbstractService {
	constructor() {
		super({
			create: '/namespaces/{namespaceId}/users',
			getAll: '/namespaces/{namespaceId}/users',
			update: '/namespaces/{namespaceId}/users/{userId}',
			delete: '/namespaces/{namespaceId}/users/{userId}',
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