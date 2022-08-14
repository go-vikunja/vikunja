import AbstractModel, { type IAbstract } from './abstractModel'
import UserModel, { type IUser } from './user'
import TeamMemberModel, { type ITeamMember } from './teamMember'
import {RIGHTS, type Right} from '@/constants/rights'

export interface ITeam extends IAbstract {
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
	id = 0
	name = ''
	description = ''
	members: ITeamMember[] = []
	right: Right = RIGHTS.READ

	createdBy: IUser = {} // FIXME: seems wrong
	created: Date = null
	updated: Date = null

	constructor(data: Partial<ITeam>) {
		super()
		this.assignData(data)

		// Make the members to usermodels
		this.members = this.members.map(m => {
			return new TeamMemberModel(m)
		})
		this.createdBy = new UserModel(this.createdBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}