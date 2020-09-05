import AbstractService from './abstractService'
import TotpModel from '../models/totp'

export default class TotpService extends AbstractService {
	urlPrefix = '/user/settings/totp'

	constructor() {
		super({})

		this.paths.get = this.urlPrefix
	}

	modelFactory(data) {
		return new TotpModel(data)
	}

	enroll() {
		return this.post(`${this.urlPrefix}/enroll`, {})
	}

	enable(model) {
		return this.post(`${this.urlPrefix}/enable`, model)
	}

	disable(model) {
		return this.post(`${this.urlPrefix}/disable`, model)
	}

	qrcode() {
		return this.http({
			url: `${this.urlPrefix}/qrcode`,
			method: 'GET',
			responseType: 'blob',
		}).then(response => {
			return Promise.resolve(new Blob([response.data]))
		})
	}
}