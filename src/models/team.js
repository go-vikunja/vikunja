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

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			id: 0,
			name: '',
			description: '',
			members: [],
			right: 0,

			createdBy: {},
			created: null,
			updated: null,
		}
	}
}