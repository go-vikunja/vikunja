import AbstractModel from './abstractModel'

export default class UserShareBaseModel extends AbstractModel {
	defaults() {
		return {
			userID: 0,
			right: 0,
			
			created: 0,
			updated: 0,
		}
	}
}