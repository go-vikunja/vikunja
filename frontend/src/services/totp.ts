import AbstractService from './abstractService'
import TotpModel from '@/models/totp'
import type {ITotp} from '@/modelTypes/ITotp'

export default class TotpService extends AbstractService<ITotp> {
	urlPrefix = '/user/settings/totp'

	constructor() {
		super({})

		this.paths.get = this.urlPrefix
	}

	modelFactory(data: Partial<ITotp>) {
		return new TotpModel(data)
	}

	enroll() {
		return this.post(`${this.urlPrefix}/enroll`, {} as ITotp)
	}

	enable(model: { passcode: string }) {
		return this.post(`${this.urlPrefix}/enable`, model as unknown as ITotp)
	}

	disable(model: { password: string }) {
		return this.post(`${this.urlPrefix}/disable`, model as unknown as ITotp)
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
