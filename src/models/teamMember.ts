import UserModel from './user'
import type ListModel from './list'

export default class TeamMemberModel extends UserModel {
	admin: boolean
	teamId: ListModel['id']

	defaults() {
		return {
			...super.defaults(),
			admin: false,
			teamId: 0,
		}
	}
}