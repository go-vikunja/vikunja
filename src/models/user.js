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
			avatarUrl: '',
			email: '',
			username: '',
			created: null,
			updated: null,
		}
	}

	getAvatarUrl(size = 50) {
		return `https://www.gravatar.com/avatar/${this.avatarUrl}?s=${size}&d=mp`
	}
}