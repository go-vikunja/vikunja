import AbstractService from './abstractService'
import TeamModel from '../models/team'
import moment from 'moment'

export default class TeamService extends AbstractService {
	constructor() {
		super({
			create: '/teams',
			get: '/teams/{id}',
			getAll: '/teams',
			update: '/teams/{id}',
			delete: '/teams/{id}',
		});
	}

	processModel(model) {
		model.created = moment(model.created).toISOString()
		model.updated = moment(model.updated).toISOString()
		return model
	}

	modelFactory(data) {
		return new TeamModel(data)
	}
}