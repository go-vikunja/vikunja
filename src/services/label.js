import AbstractService from './abstractService'
import LabelModel from '../models/label'
import {formatISO} from 'date-fns'

export default class LabelService extends AbstractService {
	constructor() {
		super({
			create: '/labels',
			getAll: '/labels',
			get: '/labels/{id}',
			update: '/labels/{id}',
			delete: '/labels/{id}',
		})
	}

	processModel(model) {
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)
		return model
	}

	modelFactory(data) {
		return new LabelModel(data)
	}
	
	beforeUpdate(label) {
		label.hex_color = label.hex_color.substring(1, 7)
		return label
	}
	
	beforeCreate(label) {
		label.hex_color = label.hex_color.substring(1, 7)
		return label
	}
}