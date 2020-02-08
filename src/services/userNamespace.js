import AbstractService from './abstractService'
import UserNamespaceModel from '../models/userNamespace'
import UserModel from '../models/user'
import moment from 'moment'

export default class UserNamespaceService extends AbstractService {
	constructor() {
		super({
			create: '/namespaces/{namespaceID}/users',
			getAll: '/namespaces/{namespaceID}/users',
			update: '/namespaces/{namespaceID}/users/{userID}',
			delete: '/namespaces/{namespaceID}/users/{userID}',
		})
	}

	processModel(model) {
		model.created = moment(model.created).toISOString()
		model.updated = moment(model.updated).toISOString()
		return model
	}

	modelFactory(data) {
		return new UserNamespaceModel(data)
	}

	modelGetAllFactory(data) {
		return new UserModel(data)
	}
}