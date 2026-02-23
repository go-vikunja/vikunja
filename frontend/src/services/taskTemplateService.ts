import AbstractService from './abstractService'
import TaskTemplateModel from '@/models/taskTemplate'
import type {ITaskTemplate} from '@/modelTypes/ITaskTemplate'

export default class TaskTemplateService extends AbstractService<ITaskTemplate> {
	constructor() {
		super({
			create: '/tasktemplates',
			get: '/tasktemplates/{id}',
			getAll: '/tasktemplates',
			update: '/tasktemplates/{id}',
			delete: '/tasktemplates/{id}',
		})
	}

	modelFactory(data) {
		return new TaskTemplateModel(data)
	}
}
