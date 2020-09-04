import AbstractService from './abstractService'
import TeamMemberModel from '../models/teamMember'
import {formatISO} from 'date-fns'

export default class TeamMemberService extends AbstractService {
	constructor() {
		super({
			create: '/teams/{teamId}/members',
			delete: '/teams/{teamId}/members/{username}',
			update: '/teams/{teamId}/members/{username}/admin',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new TeamMemberModel(data)
	}

	beforeCreate(model) {
		model.userId = model.id // The api wants to get the user id as user_Id
		model.admin = model.admin === null ? false : model.admin
		return model
	}
}