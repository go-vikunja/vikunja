import AbstractService from './abstractService'
import TeamModel, { type ITeam } from '../models/team'
import {formatISO} from 'date-fns'

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

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new TeamModel(data)
	}
}