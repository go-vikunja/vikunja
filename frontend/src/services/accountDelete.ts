import AbstractService from './abstractService'

export default class AccountDeleteService extends AbstractService {
	request(password: string) {
		return this.post('/user/deletion/request', {password})
	}
	
	confirm(token: string) {
		return this.post('/user/deletion/confirm', {token})
	}
	
	cancel(password: string) {
		return this.post('/user/deletion/cancel', {password})
	}
}
