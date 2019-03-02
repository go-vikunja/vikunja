import AbstractModel from './abstractModel'

/**
 * This class is a base class for common team sharing model.
 * It is extended in a way so it can be used for namespaces as well for lists.
 */
export default class TeamShareBaseModel extends AbstractModel {
	defaults() {
		return {
			teamID: 0,
			right: 0,
			
			created: 0,
			updated: 0
		}
	}
}