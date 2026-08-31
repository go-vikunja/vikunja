import UserShareBaseModel from './userShareBase'

import type {IUserProject} from '@/modelTypes/IUserProject'

// This class extends the user share model with a 'permissions' parameter which is used in sharing
export default class UserProjectModel extends UserShareBaseModel implements IUserProject {
	projectId = 0

	constructor(data: Partial<IUserProject>) {
		super(data)
		this.assignData(data)
	}
}
