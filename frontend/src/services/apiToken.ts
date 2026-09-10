import AbstractService from '@/services/abstractService'
import type {IApiToken} from '@/modelTypes/IApiToken'
import ApiTokenModel from '@/models/apiTokenModel'
import {toISOStringOrNull} from '@/helpers/time/toISOStringOrNull'

export type ApiTokenRoutes = Record<string, Record<string, {path: string, method: string}>>

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
			expiresAt: toISOStringOrNull(model.expiresAt),
			created: toISOStringOrNull(model.created),
		}
	}
	
	modelFactory(data: Partial<IApiToken>) {
		return new ApiTokenModel(data)
	}
	
	async getAvailableRoutes(): Promise<ApiTokenRoutes> {
		const cancel = this.setLoading()

		try {
			const response = await this.http.get('/routes')
			return response.data
		} finally {
			cancel()
		}
	}
}
