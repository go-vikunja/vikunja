import UserModel from './user'

export default class TeamMemberModel extends UserModel {
	defaults() {
		return {
			...super.defaults(),
			admin: false,
			teamId: 0,
		}
	}
}