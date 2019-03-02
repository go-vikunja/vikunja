import AbstractService from './abstractService'
import TeamModel from '../models/team'

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

	modelFactory(data) {
		return new TeamModel(data)
	}
}