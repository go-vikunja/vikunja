import AbstractService from './abstractService'
import LabelModel from '@/models/label'
import type {ILabel} from '@/modelTypes/ILabel'
import {colorFromHex} from '@/helpers/color/colorFromHex'

export default class LabelService extends AbstractService<ILabel> {
	constructor() {
		super({
			create: '/labels',
			getAll: '/labels',
			get: '/labels/{id}',
			update: '/labels/{id}',
			delete: '/labels/{id}',
		})
	}

	processModel(label: any) {
		label.created = new Date(label.created).toISOString()
		label.updated = new Date(label.updated).toISOString()
		label.hexColor = colorFromHex(label.hexColor)
		return label
	}

	modelFactory(data: Partial<ILabel>) {
		return new LabelModel(data)
	}

	beforeUpdate(label: any) {
		return this.processModel(label)
	}

	beforeCreate(label: any) {
		return this.processModel(label)
	}
}
