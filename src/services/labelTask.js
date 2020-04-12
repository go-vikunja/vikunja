import AbstractService from "./abstractService";
import LabelTask from "../models/labelTask";

export default class LabelTaskService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{taskID}/labels',
			getAll: '/tasks/{taskID}/labels',
			delete: '/tasks/{taskID}/labels/{labelId}',
		})
	}
	
	modelFactory(data) {
		return new LabelTask(data)
	}
}