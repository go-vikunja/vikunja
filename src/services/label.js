import AbstractService from './abstractService'
import LabelModel from '../models/label'
import {formatISO} from 'date-fns'
import {colorFromHex} from '@/helpers/color/colorFromHex'

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

	processModel(label) {
		label.created = formatISO(new Date(label.created))
		label.updated = formatISO(new Date(label.updated))
		label.hexColor = colorFromHex(label.hexColor)
		return label
	}

	modelFactory(data) {
		return new LabelModel(data)
	}

	beforeUpdate(label) {
		return this.processModel(label)
	}

	beforeCreate(label) {
		return this.processModel(label)
	}
}