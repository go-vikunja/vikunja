import UserModel from './user'
import type { IList } from './list'

export interface ITeamMember extends UserModel {
	admin: boolean
	teamId: IList['id']
}

export default class TeamMemberModel extends UserModel implements ITeamMember {
	admin!: boolean
	teamId!: IList['id']

	defaults() {
		return {
			...super.defaults(),
			admin: false,
			teamId: 0,
		}
	}
}