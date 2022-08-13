import AbstractModel from './abstractModel'
import UserModel, { type IUser } from './user'
import TeamMemberModel, { type ITeamMember } from './teamMember'
import {RIGHTS, type Right} from '@/constants/rights'

export interface ITeam extends AbstractModel {
	id: number
	name: string
	description: string
	members: ITeamMember[]
	right: Right

	createdBy: IUser
	created: Date
	updated: Date
}

export default class TeamModel extends AbstractModel implements ITeam {
	declare id: number
	declare name: string
	declare description: string
	members: ITeamMember[]
	declare right: Right

	createdBy: IUser
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