import AbstractService from './abstractService'
import NamespaceModel from '../models/namespace'
import {formatISO} from 'date-fns'

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

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new NamespaceModel(data)
	}

	beforeUpdate(namespace) {
		namespace.hexColor = namespace.hexColor.substring(1, 7)
		return namespace
	}

	beforeCreate(namespace) {
		namespace.hexColor = namespace.hexColor.substring(1, 7)
		return namespace
	}
}