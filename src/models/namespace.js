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

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	// Default attributes that define the 'empty' state.
	defaults() {
		return {
			id: 0,
			name: '',
			description: '',
			owner: UserModel,
			lists: [],

			created: null,
			updated: null,
		}
	}
}