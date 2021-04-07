import AbstractModel from './abstractModel'
import UserModel from './user'

export default class ListModel extends AbstractModel {

	constructor(data) {
		// The constructor of AbstractModel handles all the default parsing.
		super(data)

		this.sharedBy = new UserModel(this.sharedBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	// Default attributes that define the "empty" state.
	defaults() {
		return {
			id: 0,
			hash: '',
			right: 0,
			sharedBy: UserModel,
			sharingType: 0,
			listId: 0,
			name: '',

			created: null,
			updated: null,
		}
	}
}