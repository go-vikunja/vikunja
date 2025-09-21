import AbstractService from './abstractService'

export default class AccountDeleteService extends AbstractService {
	request(password: string) {
		return this.post('/user/deletion/request', {password, maxPermission: null})
	}

	confirm(token: string) {
		return this.post('/user/deletion/confirm', {token, maxPermission: null})
	}

	cancel(password: string) {
		return this.post('/user/deletion/cancel', {password, maxPermission: null})
	}
}
