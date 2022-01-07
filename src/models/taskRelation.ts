import AbstractModel from './abstractModel'
import UserModel from './user'

export default class TaskRelationModel extends AbstractModel {
	constructor(data) {
		super(data)
		this.createdBy = new UserModel(this.createdBy)
		this.created = new Date(this.created)
	}

	defaults() {
		return {
			id: 0,
			otherTaskId: 0,
			taskId: 0,
			relationKind: '',

			createdBy: UserModel,
			created: null,
		}
	}
}