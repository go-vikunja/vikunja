import AbstractModel from './abstractModel'
import type TeamModel from './team'
import {RIGHTS, type Right} from '@/models/constants/rights'

/**
 * This class is a base class for common team sharing model.
 * It is extended in a way so it can be used for namespaces as well for lists.
 */
export default class TeamShareBaseModel extends AbstractModel {
	teamId: TeamModel['id']
	right: Right

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