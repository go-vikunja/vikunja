import AbstractModel from './abstractModel'
import UserModel from './user'
import TeamMemberModel from './teamMember'
import {RIGHTS, type Right} from '@/models/constants/rights'

export default class TeamModel extends AbstractModel {
	id: 0
	name: string
	description: string
	members: TeamMemberModel[]
	right: Right

	createdBy: UserModel
	created: Date
	updated: Date

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
			right: RIGHTS.READ,

			createdBy: {},
			created: null,
			updated: null,
		}
	}
}