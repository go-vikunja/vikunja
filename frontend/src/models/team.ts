import AbstractModel from './abstractModel'
import UserModel from './user'
import TeamMemberModel from './teamMember'

import {PERMISSIONS, type Permission} from '@/constants/permissions'
import type {ITeam} from '@/modelTypes/ITeam'
import type {ITeamMember} from '@/modelTypes/ITeamMember'
import type {IUser} from '@/modelTypes/IUser'

export default class TeamModel extends AbstractModel<ITeam> implements ITeam {
	id = 0
	name = ''
	description = ''
	members: ITeamMember[] = []
	permission: Permission = PERMISSIONS.READ
	externalId = ''
	isPublic: boolean = false

	createdBy: IUser = new UserModel()
	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<ITeam> = {}) {
		super()
		this.assignData(data)

		// Make the members to usermodels
		this.members = this.members.map(m => {
			return new TeamMemberModel(m)
		})
		this.createdBy = this.createdBy ? new UserModel(this.createdBy) : new UserModel()

		this.created = this.created ? new Date(this.created) : new Date()
		this.updated = this.updated ? new Date(this.updated) : new Date()
	}
}
