import AbstractModel, { type IAbstract } from './abstractModel'
import ListModel, { type IList } from './list'
import type { INamespace } from './namespace'

export interface IListDuplicate extends IAbstract {
	listId: number
	namespaceId: INamespace['id']
	list: IList
}

export default class ListDuplicateModel extends AbstractModel implements IListDuplicate {
	listId!: number
	namespaceId!: INamespace['id']
	list: IList

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