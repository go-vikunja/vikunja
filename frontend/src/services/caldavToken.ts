import CaldavTokenModel from '@/models/caldavToken'
import type {ICaldavToken} from '@/modelTypes/ICaldavToken'
import AbstractService from './abstractService'

export default class CaldavTokenService extends AbstractService<ICaldavToken> {
	constructor() {
		super({
			getAll: '/user/settings/token/caldav',
			create: '/user/settings/token/caldav',
			delete: '/user/settings/token/caldav/{id}',
		})
	}

	modelFactory(data) {
		return new CaldavTokenModel(data)
	}
}
