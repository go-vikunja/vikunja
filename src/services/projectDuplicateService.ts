import AbstractService from './abstractService'
import listDuplicateModel from '@/models/listDuplicateModel'
import type {IListDuplicate} from '@/modelTypes/IListDuplicate'

export default class ListDuplicateService extends AbstractService<IListDuplicate> {
	constructor() {
		super({
			create: '/lists/{listId}/duplicate',
		})
	}

	beforeCreate(model) {

		model.list = null
		return model
	}

	modelFactory(data) {
		return new listDuplicateModel(data)
	}
}