import AbstractModel from './abstractModel'
import ListModel from './list'
import UserModel from './user'

export default class NamespaceModel extends AbstractModel {
	constructor(data) {
		super(data)

		if (this.hex_color !== '' && this.hex_color.substring(0, 1) !== '#') {
			this.hex_color = '#' + this.hex_color
		}

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
			is_archived: false,
			hex_color: '',

			created: null,
			updated: null,
		}
	}
}