import {formatISO} from 'date-fns'
import CaldavTokenModel from '../models/caldavToken'
import AbstractService from './abstractService'

export default class CaldavTokenService extends AbstractService {
	constructor() {
		super({
			getAll: '/user/settings/token/caldav',
			create: '/user/settings/token/caldav',
			delete: '/user/settings/token/caldav/{id}',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		return model
	}

	modelFactory(data) {
		return new CaldavTokenModel(data)
	}
}
	