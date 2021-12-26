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
		return {
			...model,
			created: formatISO(new Date(model.created)),
		}
	}

	modelFactory(data) {
		return new CaldavTokenModel(data)
	}
}
	