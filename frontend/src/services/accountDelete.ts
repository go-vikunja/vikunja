import AbstractService from './abstractService'

export default class AccountDeleteService extends AbstractService {
	request(password: string) {
		return this.post('/user/deletion/request', {password} as any)
	}
	
	confirm(token: string) {
		return this.post('/user/deletion/confirm', {token} as any)
	}
	
	cancel(password: string) {
		return this.post('/user/deletion/cancel', {password} as any)
	}
}
