import AbstractService from './abstractService'
import LabelModel from '../models/label'

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