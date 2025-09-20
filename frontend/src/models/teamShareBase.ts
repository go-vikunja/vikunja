import AbstractModel from './abstractModel'

import {PERMISSIONS, type Permission} from '@/constants/permissions'
import type {ITeamShareBase} from '@/modelTypes/ITeamShareBase'
import type {ITeam} from '@/modelTypes/ITeam'

/**
 * This class is a base class for common team sharing model.
 * It is extended in a way, so it can be used for projects.
 */
export default class TeamShareBaseModel extends AbstractModel<ITeamShareBase> implements ITeamShareBase {
	teamId: ITeam['id'] = 0
	permission: Permission = PERMISSIONS.READ

	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<ITeamShareBase>) {
		super()
		this.assignData(data)

		this.created = this.created ? new Date(this.created) : new Date()
		this.updated = this.updated ? new Date(this.updated) : new Date()
	}
}
