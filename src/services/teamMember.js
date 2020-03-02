import AbstractService from './abstractService'
import TeamMemberModel from '../models/teamMember'
import {formatISO} from 'date-fns'

export default class TeamMemberService extends AbstractService {
	constructor() {
		super({
			create: '/teams/{teamID}/members',
			delete: '/teams/{teamID}/members/{id}', // "id" is the user id because we're intheriting from a normal user
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
		model.userID = model.id // The api wants to get the user id as userID
		model.admin = model.admin === null ? false : model.admin
		return model
	}
}