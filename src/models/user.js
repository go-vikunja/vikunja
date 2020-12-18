import AbstractModel from './abstractModel'

export default class UserModel extends AbstractModel {
	constructor(data) {
		super(data)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			id: 0,
			email: '',
			username: '',
			name: '',
			created: null,
			updated: null,
		}
	}

	getAvatarUrl(size = 50) {
		return `${window.API_URL}/avatar/${this.username}?size=${size}`
	}

	getDisplayName() {
		if (this.name !== '') {
			return this.name
		}

		return this.username
	}
}