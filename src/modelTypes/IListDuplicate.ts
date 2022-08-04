import type {IAbstract} from './IAbstract'
import type {IList} from './IList'
import type {INamespace} from './INamespace'

export interface IListDuplicate extends IAbstract {
	listId: number
	namespaceId: INamespace['id']
	list: IList
}