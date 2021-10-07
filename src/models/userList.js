import UserShareBaseModel from './userShareBase'

// This class extends the user share model with a 'rights' parameter which is used in sharing
export default class UserListModel extends UserShareBaseModel {
	defaults() {
		return {
			...super.defaults(),
			listId: 0,
		}
	}
}