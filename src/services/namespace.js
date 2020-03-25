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
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)
		return model
	}

	modelFactory(data) {
		return new NamespaceModel(data)
	}

	beforeUpdate(namespace) {
		namespace.hex_color = namespace.hex_color.substring(1, 7)
		return namespace
	}

	beforeCreate(namespace) {
		namespace.hex_color = namespace.hex_color.substring(1, 7)
		return namespace
	}
}