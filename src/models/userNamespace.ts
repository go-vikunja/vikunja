import UserShareBaseModel from './userShareBase'
import type { INamespace } from './namespace'

export interface IUserNamespace extends UserShareBaseModel {
	namespaceId: INamespace['id']
}

// This class extends the user share model with a 'rights' parameter which is used in sharing
export default class UserNamespaceModel extends UserShareBaseModel implements IUserNamespace {
	namespaceId: INamespace['id'] = 0

	constructor(data: Partial<IUserNamespace>) {
		super(data)
		this.assignData(data)
	}
}