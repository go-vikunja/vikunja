import AbstractService from './abstractService'
import TaskFromTemplateModel from '@/models/taskFromTemplate'
import type {ITaskFromTemplate} from '@/modelTypes/ITaskFromTemplate'

export default class TaskFromTemplateService extends AbstractService<ITaskFromTemplate> {
	constructor() {
		super({
			create: '/tasktemplates/{templateId}/tasks',
		})
	}

	modelFactory(data) {
		return new TaskFromTemplateModel(data)
	}
}
