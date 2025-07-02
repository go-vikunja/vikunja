import AbstractService from './abstractService'
import type {IMessage} from '@/modelTypes/IMessage'

export default class AccountDeleteService extends AbstractService<IMessage> {
	async request(password: string): Promise<IMessage> {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post('/user/deletion/request', {password})
			return this.modelUpdateFactory(response.data)
		} finally {
			cancel()
		}
	}
	
	async confirm(token: string): Promise<IMessage> {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post('/user/deletion/confirm', {token})
			return this.modelUpdateFactory(response.data)
		} finally {
			cancel()
		}
	}
	
	async cancel(password: string): Promise<IMessage> {
		const cancel = this.setLoading()
		try {
			const response = await this.http.post('/user/deletion/cancel', {password})
			return this.modelUpdateFactory(response.data)
		} finally {
			cancel()
		}
	}
}
