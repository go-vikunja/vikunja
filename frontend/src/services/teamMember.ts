import AbstractService from './abstractService'
import TeamMemberModel from '@/models/teamMember'
import type {ITeamMember} from '@/modelTypes/ITeamMember'

export default class TeamMemberService extends AbstractService<ITeamMember> {
	constructor() {
		super({
			create: '/teams/{teamId}/members',
			delete: '/teams/{teamId}/members/{username}',
			update: '/teams/{teamId}/members/{username}/admin',
		})
	}

	modelFactory(data: Partial<ITeamMember>): ITeamMember {
		return new TeamMemberModel(data)
	}

	beforeCreate(model: ITeamMember): ITeamMember {
		// The api wants to get the user id as user_Id
		const modelWithUserId = model as ITeamMember & { userId: number }
		modelWithUserId.userId = model.id
		modelWithUserId.admin = model.admin === null ? false : model.admin
		return modelWithUserId
	}
}
