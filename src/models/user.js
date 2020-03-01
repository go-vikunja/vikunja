import AbstractModel from './abstractModel'
import config from '../../public/config'

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
			created: null,
			updated: null,
		}
	}

	getAvatarUrl(size = 50) {
		return `${config.VIKUNJA_API_BASE_URL}${this.username}/avatar?size=${size}`
	}
}