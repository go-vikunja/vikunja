import UserShareBaseModel from './userShareBase'
import type { IList } from './list'

export interface IUserList extends UserShareBaseModel {
	listId: IList['id']
}

// This class extends the user share model with a 'rights' parameter which is used in sharing
export default class UserListModel extends UserShareBaseModel implements IUserList {
	declare listId: IList['id']

	defaults() {
		return {
			...super.defaults(),
			listId: 0,
		}
	}
}