import {formatISO} from 'date-fns'
import CaldavTokenModel, {type ICaldavToken} from '../models/caldavToken'
import AbstractService from './abstractService'

export default class CaldavTokenService extends AbstractService<ICaldavToken> {
	constructor() {
		super({
			getAll: '/user/settings/token/caldav',
			create: '/user/settings/token/caldav',
			delete: '/user/settings/token/caldav/{id}',
		})
	}

	processModel(model: Partial<ICaldavToken>) {
		return {
			...model,
			created: formatISO(new Date(model.created)),
		}
	}

	modelFactory(data) {
		return new CaldavTokenModel(data)
	}
}
	