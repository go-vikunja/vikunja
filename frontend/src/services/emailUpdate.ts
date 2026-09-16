import AbstractService from './abstractService'
import {apiV2Url} from '@/helpers/fetcher'
import type {IEmailUpdate} from '@/modelTypes/IEmailUpdate'

export default class EmailUpdateService extends AbstractService<IEmailUpdate> {
	// The email update endpoints only exist on /api/v2, hence the absolute URLs.
	async update(model: IEmailUpdate) {
		const cancel = this.setLoading()
		try {
			// v2 rejects unknown properties, so never send the whole model (AbstractModel adds maxPermission)
			await this.http.put(apiV2Url('user/settings/email'), {
				newEmail: model.newEmail,
				password: model.password,
			})
			return model
		} finally {
			cancel()
		}
	}

	async cancel() {
		const cancel = this.setLoading()
		try {
			await this.http.delete(apiV2Url('user/settings/email'))
		} finally {
			cancel()
		}
	}

	async resend() {
		const cancel = this.setLoading()
		try {
			await this.http.post(apiV2Url('user/settings/email/resend'))
		} finally {
			cancel()
		}
	}
}
