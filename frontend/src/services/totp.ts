import AbstractService from './abstractService'
import TotpModel from '@/models/totp'
import type {ITotp} from '@/modelTypes/ITotp'

export default class TotpService extends AbstractService<ITotp> {
	urlPrefix = '/user/settings/totp'

	constructor() {
		super({})

		this.paths.get = this.urlPrefix
	}

	modelFactory(data: Partial<ITotp>): ITotp {
		return new TotpModel(data) as ITotp
	}

	enroll(): Promise<any> {
		return this.post(`${this.urlPrefix}/enroll`, {})
	}

	enable(model: ITotp): Promise<any> {
		return this.post(`${this.urlPrefix}/enable`, model)
	}

	disable(model: ITotp): Promise<any> {
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
