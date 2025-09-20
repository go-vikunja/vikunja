import AbstractService from './abstractService'
import PasswordResetModel from '@/models/passwordReset'
import type {IPasswordReset} from '@/modelTypes/IPasswordReset'

interface PasswordResetPaths {
	reset: string
	requestReset: string
}

export default class PasswordResetService extends AbstractService<IPasswordReset> {
	declare paths: PasswordResetPaths

	constructor() {
		super({})
		this.paths = {
			reset: '/user/password/reset',
			requestReset: '/user/password/token',
		}
	}

	modelFactory(data: Partial<IPasswordReset>) {
		return new PasswordResetModel(data)
	}

	async resetPassword(model: IPasswordReset) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post(this.paths.reset, model)
			return this.modelFactory(response.data)
		} finally {
			cancel()
		}
	}

	async requestResetPassword(model: Partial<IPasswordReset>) {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post(this.paths.requestReset, model)
			return this.modelFactory(response.data)
		} finally {
			cancel()
		}
	}
}
