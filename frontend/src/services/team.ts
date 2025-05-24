import AbstractService from './abstractService'
import TeamModel from '@/models/team'
import type {ITeam} from '@/modelTypes/ITeam'

export default class TeamService extends AbstractService<ITeam> {
	constructor() {
		super({
			create: '/teams',
			get: '/teams/{id}',
			getAll: '/teams',
			update: '/teams/{id}',
			delete: '/teams/{id}',
		})
	}

	modelFactory(data) {
		return new TeamModel(data)
	}
}
