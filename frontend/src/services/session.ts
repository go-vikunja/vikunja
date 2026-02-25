import AbstractService from '@/services/abstractService'
import SessionModel from '@/models/session'
import type {ISession} from '@/modelTypes/ISession'

export default class SessionService extends AbstractService<ISession> {
	constructor() {
		super({
			getAll: '/user/sessions',
			delete: '/user/sessions/{id}',
		})
	}

	modelFactory(data: Partial<ISession>) {
		return new SessionModel(data)
	}
}
