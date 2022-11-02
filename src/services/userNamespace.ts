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

	modelFactory(data) {
		return new UserNamespaceModel(data)
	}

	modelGetAllFactory(data) {
		return new UserModel(data)
	}
}