import AbstractModel from './abstractModel'
import {RIGHTS, type Right} from '@/constants/rights'
import type { ITeam } from './team'

export interface ITeamShareBase extends AbstractModel {
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
	declare teamId: ITeam['id']
	declare right: Right

	created: Date
	updated: Date

	constructor(data) {
		super(data)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			teamId: 0,
			right: RIGHTS.READ,

			created: null,
			updated: null,
		}
	}
}