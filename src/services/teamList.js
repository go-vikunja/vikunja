import AbstractService from './abstractService'
import TeamListModel from '../models/teamList'
import TeamModel from '../models/team'
import {formatISO} from 'date-fns'

export default class TeamListService extends AbstractService {
	constructor() {
		super({
			create: '/lists/{listID}/teams',
			getAll: '/lists/{listID}/teams',
			update: '/lists/{listID}/teams/{teamID}',
			delete: '/lists/{listID}/teams/{teamID}',
		})
	}

	processModel(model) {
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)
		return model
	}

	modelFactory(data) {
		return new TeamListModel(data)
	}

	modelGetAllFactory(data) {
		return new TeamModel(data)
	}
}