import UserModel from './user'

import type {ITeamMember} from '@/modelTypes/ITeamMember'
import type {IList} from '@/modelTypes/IList'

export default class TeamMemberModel extends UserModel implements ITeamMember {
	admin = false
	teamId: IList['id'] = 0

	constructor(data: Partial<ITeamMember>) {
		super(data)
		this.assignData(data)
	}
}