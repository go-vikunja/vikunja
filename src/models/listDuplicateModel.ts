import AbstractModel from './abstractModel'
import ListModel from './list'

export default class ListDuplicateModel extends AbstractModel {
	constructor(data) {
		super(data)
		this.list = new ListModel(this.list)
	}

	defaults() {
		return {
			listId: 0,
			namespaceId: 0,
			list: ListModel,
		}
	}
}