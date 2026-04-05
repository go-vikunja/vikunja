import AbstractService from '@/services/abstractService'
import type {IApiToken} from '@/modelTypes/IApiToken'
import ApiTokenModel from '@/models/apiTokenModel'

export default class BotTokenService extends AbstractService<IApiToken> {
	constructor(botId: number) {
		super({
			create: `/user/bots/${botId}/tokens`,
			getAll: `/user/bots/${botId}/tokens`,
			delete: `/user/bots/${botId}/tokens/{id}`,
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
}
