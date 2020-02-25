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
			avatar: '',
			email: '',
			username: '',
			created: null,
			updated: null,
		}
	}

	getAvatarUrl(size = 50) {
		const avatarUrl = this.avatar !== '' ? this.avatar : this.avatarUrl
		return `https://www.gravatar.com/avatar/${avatarUrl}?s=${size}&d=mp`
	}
}