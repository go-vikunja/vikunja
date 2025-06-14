import AbstractService from './abstractService'
import TotpModel from '@/models/totp'
import type {ITotp} from '@/modelTypes/ITotp'

export default class TotpService extends AbstractService<ITotp> {
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

	async qrcode() {
		const response = await this.http({
			url: `${this.urlPrefix}/qrcode`,
			method: 'GET',
			responseType: 'blob',
		})
		return new Blob([response.data])
	}
}
