import AbstractModel from './abstractModel'
import ListModel from './list'
import UserModel from './user'

export default class NamespaceModel extends AbstractModel {
	constructor(data) {
		super(data)

		this.lists = this.lists.map(l => {
			return new ListModel(l)
		})
		this.owner = new UserModel(this.owner)
	}

	// Default attributes that define the 'empty' state.
	defaults() {
		return {
			id: 0,
			name: '',
			description: '',
			owner: UserModel,
			lists: [],

			created: 0,
			updated: 0,
		}
	}
}