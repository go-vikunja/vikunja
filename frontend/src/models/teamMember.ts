import UserModel from './user'

import type {ITeamMember} from '@/modelTypes/ITeamMember'

export default class TeamMemberModel extends UserModel implements ITeamMember {
	admin = false
	teamId = 0

	constructor(data: Partial<ITeamMember>) {
		super(data)
		this.assignData(data)
	}
}
