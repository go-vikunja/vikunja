import UserModel from './user'

import type {ITeamMember} from '@/modelTypes/ITeamMember'
import type {IProject} from '@/modelTypes/IProject'

export default class TeamMemberModel extends UserModel implements ITeamMember {
	admin = false
	teamId: IProject['id'] = 0

	constructor(data: Partial<ITeamMember>) {
		super(data)
		this.assignData(data)
	}
}
