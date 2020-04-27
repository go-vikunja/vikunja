import AbstractService from './abstractService'
import TeamMemberModel from '../models/teamMember'
import {formatISO} from 'date-fns'

export default class TeamMemberService extends AbstractService {
	constructor() {
		super({
			create: '/teams/{teamId}/members',
			delete: '/teams/{teamId}/members/{username}',
		});
	}

	processModel(model) {
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)
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