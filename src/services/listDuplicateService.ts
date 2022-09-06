import AbstractService from './abstractService'
import listDuplicateModel, {type IListDuplicate} from '../models/listDuplicateModel'

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