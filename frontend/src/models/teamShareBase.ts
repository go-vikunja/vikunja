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

	created: Date = null
	updated: Date = null

	constructor(data: Partial<ITeamShareBase>) {
		super()
		this.assignData(data)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
