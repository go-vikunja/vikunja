import AbstractModel from './abstractModel'
import ListModel from './list'

import type {IListDuplicate} from '@/modelTypes/IListDuplicate'
import type {INamespace} from '@/modelTypes/INamespace'
import type {IList} from '@/modelTypes/IList'

export default class ListDuplicateModel extends AbstractModel<IListDuplicate> implements IListDuplicate {
	listId = 0
	namespaceId: INamespace['id'] = 0
	list: IList = ListModel

	constructor(data : Partial<IListDuplicate>) {
		super()
		this.assignData(data)

		this.list = new ListModel(this.list)
	}
}