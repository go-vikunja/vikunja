import AbstractModel from './abstractModel'
import ListModel from './list'
import NamespaceModel from './namespace'

export default class ListDuplicateModel extends AbstractModel {
	listId: number
	namespaceId: NamespaceModel['id']
	list: ListModel

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