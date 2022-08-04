import type {IUserShareBase} from './IUserShareBase'
import type {IList} from './IList'

export interface IUserList extends IUserShareBase {
	listId: IList['id']
}