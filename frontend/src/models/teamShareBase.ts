import AbstractModel from './abstractModel'

import {RIGHTS, type Right} from '@/constants/rights'
import type {ITeamShareBase} from '@/modelTypes/ITeamShareBase'
import type {ITeam} from '@/modelTypes/ITeam'

/**
 * This class is a base class for common team sharing model.
 * It is extended in a way, so it can be used for projects.
 */
export default class TeamShareBaseModel extends AbstractModel<ITeamShareBase> implements ITeamShareBase {
	teamId: ITeam['id'] = 0
	right: Right = RIGHTS.READ

	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<ITeamShareBase>) {
		super()
		this.assignData(data)

		this.created = new Date(this.created || Date.now())
		this.updated = new Date(this.updated || Date.now())
	}
}
