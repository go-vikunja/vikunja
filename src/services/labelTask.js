import AbstractService from "./abstractService";
import LabelTask from "../models/labelTask";

export default class LabelTaskService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{taskId}/labels',
			getAll: '/tasks/{taskId}/labels',
			delete: '/tasks/{taskId}/labels/{labelId}',
		})
	}
	
	modelFactory(data) {
		return new LabelTask(data)
	}
}