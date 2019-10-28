import AbstractModel from './abstractModel'
import UserModel from "./user";

export default class TaskRelationModel extends AbstractModel {
	constructor(data) {
		super(data)
		this.created_by = new UserModel(this.created_by)
	}
	
	defaults() {
		return {
			id: 0,
			other_task_id: 0,
			task_id: 0,
			relation_kind: '',

			created_by: UserModel,
			created: 0,
		}
	}
}