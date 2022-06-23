import UserShareBaseModel from './userShareBase'
import type NamespaceModel from './namespace'

// This class extends the user share model with a 'rights' parameter which is used in sharing
export default class UserNamespaceModel extends UserShareBaseModel {
	namespaceId: NamespaceModel['id']

	defaults() {
		return {
			...super.defaults(),
			namespaceId: 0,
		}
	}
}