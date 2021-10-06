import UserShareBaseModel from './userShareBase'

// This class extends the user share model with a 'rights' parameter which is used in sharing
export default class UserNamespaceModel extends UserShareBaseModel {
	defaults() {
		return {
			...super.defaults(),
			namespaceId: 0,
		}
	}
}