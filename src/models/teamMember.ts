import UserModel from './user'
import type { IList } from './list'

export interface ITeamMember extends UserModel {
	admin: boolean
	teamId: IList['id']
}

export default class TeamMemberModel extends UserModel implements ITeamMember {
	admin = false
	teamId: IList['id'] = 0

	constructor(data: Partial<ITeamMember>) {
		super(data)
		this.assignData(data)
	}
}