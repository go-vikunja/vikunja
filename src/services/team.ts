import AbstractService from './abstractService'
import TeamModel from '@/models/team'
import type { ITeam } from '@/modelTypes/ITeam'
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