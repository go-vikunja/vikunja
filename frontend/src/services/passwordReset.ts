import AbstractService from './abstractService'
import PasswordResetModel from '@/models/passwordReset'
import type {IPasswordReset} from '@/modelTypes/IPasswordReset'

export default class PasswordResetService extends AbstractService<IPasswordReset> {

	constructor() {
		super({})
		// Note: Custom paths not available in base Paths interface
	}

	modelFactory(data: any) {
		return new PasswordResetModel(data)
	}

	async resetPassword(model: any) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post('/user/password/reset', model)
			return this.modelFactory(response.data)
		} finally {
			cancel()
		}
	}

	async requestResetPassword(model: any) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post('/user/password/token', model)
			return this.modelFactory(response.data)
		} finally {
			cancel()
		}
	}
}
