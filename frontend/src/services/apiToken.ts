import AbstractService from '@/services/abstractService'
import type {IApiToken} from '@/modelTypes/IApiToken'
import ApiTokenModel from '@/models/apiTokenModel'
import {objectToSnakeCase} from '@/helpers/case'

export default class ApiTokenService extends AbstractService<IApiToken> {
	constructor() {
		super({
			create: '/tokens',
			getAll: '/tokens',
			delete: '/tokens/{id}',
		})
	}

	// Disable the default snake_case transform — beforeCreate handles it
	// manually to preserve the permissions map keys (e.g. "time-entries").
	autoTransformBeforePut(): boolean {
		return false
	}

	beforeCreate(model: IApiToken) {
		const permissions = model.permissions
		const transformed = objectToSnakeCase(model)
		transformed.permissions = permissions
		return transformed
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
