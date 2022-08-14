import AbstractModel, { type IAbstract } from './abstractModel'
import ListModel, { type IList } from './list'
import type { INamespace } from './namespace'

export interface IListDuplicate extends IAbstract {
	listId: number
	namespaceId: INamespace['id']
	list: IList
}

export default class ListDuplicateModel extends AbstractModel implements IListDuplicate {
	listId = 0
	namespaceId: INamespace['id'] = 0
	list: IList = ListModel

	constructor(data : Partial<IListDuplicate>) {
		super()
		this.assignData(data)

		this.list = new ListModel(this.list)
	}
}