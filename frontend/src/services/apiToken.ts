import AbstractService from '@/services/abstractService'
import type {IApiToken} from '@/modelTypes/IApiToken'
import ApiTokenModel from '@/models/apiTokenModel'

export default class ApiTokenService extends AbstractService<IApiToken> {
	constructor() {
		super({
			create: '/tokens',
			getAll: '/tokens',
			delete: '/tokens/{id}',
		})
	}

	processModel(model: IApiToken) {
		return {
			...model,
			expiresAt: new Date(model.expiresAt).toISOString(),
			created: new Date(model.created).toISOString(),
		}
	}
	
	modelFactory(data: Partial<IApiToken>) {
		return new ApiTokenModel(data)
	}
	
	async getAvailableRoutes() {
		const cancel = this.setLoading()

		try {
			const response = await this.http.get('/routes')
			return response.data
		} finally {
			cancel()
		}
	}
}
