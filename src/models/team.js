import AbstractModel from './abstractModel'
import UserModel from './user'
import TeamMemberModel from './teamMember'

export default class TeamModel extends AbstractModel {
	constructor(data) {
		super(data)
		
		// Make the members to usermodels
		this.members = this.members.map(m => {
			return new TeamMemberModel(m)
		})
		this.createdBy = new UserModel(this.createdBy)
	}

	defaults() {
		return {
			id: 0,
			name: '',
			description: '',
			members: [],
			right: 0,

			createdBy: {},
			created: 0,
			updated: 0
		}
	}
}