import AbstractService from './abstractService'
import NamespaceModel from '../models/namespace'

export default class NamespaceService extends AbstractService {
	constructor() {
		super({
			create: '/namespaces',
			get: '/namespaces/{id}',
			getAll: '/namespaces',
			update: '/namespaces/{id}',
			delete: '/namespaces/{id}',
		});
	}

	modelFactory(data) {
		return new NamespaceModel(data)
	}
}