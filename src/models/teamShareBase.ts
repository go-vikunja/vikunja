import AbstractModel, { type IAbstract } from './abstractModel'
import {RIGHTS, type Right} from '@/constants/rights'
import type { ITeam } from './team'

export interface ITeamShareBase extends IAbstract {
	teamId: ITeam['id']
	right: Right

	created: Date
	updated: Date
}

/**
 * This class is a base class for common team sharing model.
 * It is extended in a way so it can be used for namespaces as well for lists.
 */
export default class TeamShareBaseModel extends AbstractModel implements ITeamShareBase {
	teamId: ITeam['id'] = 0
	right: Right = RIGHTS.READ

	created: Date = null
	updated: Date = null

	constructor(data: Partial<ITeamShareBase>) {
		super()
		this.assignData(data)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}