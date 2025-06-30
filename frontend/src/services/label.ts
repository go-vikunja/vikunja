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

	processModel(label: ILabel) {
		label.created = new Date(label.created)
		label.updated = new Date(label.updated)
		label.hexColor = colorFromHex(label.hexColor)
		return label
	}

	modelFactory(data: Partial<ILabel>) {
		return new LabelModel(data)
	}

	beforeUpdate(label: ILabel) {
		return this.processModel(label)
	}

	beforeCreate(label: ILabel) {
		return this.processModel(label)
	}
}
