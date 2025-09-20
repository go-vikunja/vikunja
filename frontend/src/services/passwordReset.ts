import AbstractService from './abstractService'
import PasswordResetModel from '@/models/passwordReset'
import type {IPasswordReset} from '@/modelTypes/IPasswordReset'

export default class PasswordResetService extends AbstractService<IPasswordReset> {

	constructor() {
		super({})
		this.paths = {
			...this.paths,
			reset: '/user/password/reset',
		} as any
		;(this.paths as any).requestReset = '/user/password/token'
	}

	modelFactory(data: Partial<IPasswordReset>) {
		return new PasswordResetModel(data)
	}

	async resetPassword(model: IPasswordReset) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post((this.paths as any).reset, model)
			return this.modelFactory(response.data)
		} finally {
			cancel()
		}
	}

	async requestResetPassword(model: Partial<IPasswordReset>) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post((this.paths as any).requestReset, model)
			return this.modelFactory(response.data)
		} finally {
			cancel()
		}
	}
}
