import AbstractModel from './abstractModel'
import UserModel from './user'
import TeamMemberModel from './teamMember'

import {RIGHTS, type Right} from '@/constants/rights'
import type {ITeam} from '@/modelTypes/ITeam'
import type {ITeamMember} from '@/modelTypes/ITeamMember'
import type {IUser} from '@/modelTypes/IUser'

export default class TeamModel extends AbstractModel<ITeam> implements ITeam {
	id = 0
	name = ''
	description = ''
	members: ITeamMember[] = []
	right: Right = RIGHTS.READ
	externalId = ''
	isPublic: boolean = false

	createdBy: IUser | null = null
	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<ITeam> = {}) {
		super()
		this.assignData(data)

		// Make the members to usermodels
		this.members = this.members.map(m => {
			return new TeamMemberModel(m)
		})
		if (this.createdBy) {
			this.createdBy = new UserModel(this.createdBy)
		}

		this.created = new Date(this.created || Date.now())
		this.updated = new Date(this.updated || Date.now())
	}
}
