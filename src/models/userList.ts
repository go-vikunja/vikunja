import UserShareBaseModel from './userShareBase'
import type ListModel from './list'
// This class extends the user share model with a 'rights' parameter which is used in sharing
export default class UserListModel extends UserShareBaseModel {
	listId: ListModel['id']

	defaults() {
		return {
			...super.defaults(),
			listId: 0,
		}
	}
}