import {formatISO} from 'date-fns'

import AbstractService from './abstractService'
import UserNamespaceModel from '@/models/userNamespace'
import type {IUserNamespace} from '@/modelTypes/IUserNamespace'
import UserModel from '@/models/user'

export default class UserNamespaceService extends AbstractService<IUserNamespace> {
	constructor() {
		super({
			create: '/namespaces/{namespaceId}/users',
			getAll: '/namespaces/{namespaceId}/users',
			update: '/namespaces/{namespaceId}/users/{userId}',
			delete: '/namespaces/{namespaceId}/users/{userId}',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new UserNamespaceModel(data)
	}

	modelGetAllFactory(data) {
		return new UserModel(data)
	}
}