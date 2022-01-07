import AbstractService from './abstractService'

export default class AccountDeleteService extends AbstractService {
	request(password) {
		return this.post('/user/deletion/request', {password: password})
	}
	
	confirm(token) {
		return this.post('/user/deletion/confirm', {token: token})
	}
	
	cancel(password) {
		return this.post('/user/deletion/cancel', {password: password})
	}
}