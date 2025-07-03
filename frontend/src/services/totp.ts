import AbstractService from './abstractService'
import TotpModel from '@/models/totp'
import type {ITotp} from '@/modelTypes/ITotp'
import type {ITotpPasscode} from '@/modelTypes/ITotpPasscode'
import type {ITotpDisable} from '@/modelTypes/ITotpDisable'

export default class TotpService extends AbstractService<ITotp> {
	urlPrefix = '/user/settings/totp'

	constructor() {
		const urlPrefix = '/user/settings/totp'
		super({get: urlPrefix})

		this.urlPrefix = urlPrefix
		this.paths.get = this.urlPrefix
	}

	modelFactory(data: Partial<ITotp>): ITotp {
		return new TotpModel(data) as ITotp
	}

	enroll(): Promise<ITotp> {
		return this.post(`${this.urlPrefix}/enroll`, new TotpModel())
	}

	enable(passcode: ITotpPasscode): Promise<{message: string}> {
		return this.post(`${this.urlPrefix}/enable`, passcode as unknown as ITotp) as unknown as Promise<{message: string}>
	}

	disable(credentials: ITotpDisable): Promise<{message: string}> {
		return this.post(`${this.urlPrefix}/disable`, credentials as unknown as ITotp) as unknown as Promise<{message: string}>
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
