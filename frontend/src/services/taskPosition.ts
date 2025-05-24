import AbstractService from '@/services/abstractService'
import type {ITaskPosition} from '@/modelTypes/ITaskPosition'
import TaskPositionModel from '@/models/taskPosition'

export default class TaskPositionService extends AbstractService<ITaskPosition> {
	constructor() {
		super({
			update: '/tasks/{taskId}/position',
		})
	}
	
	modelFactory(data: Partial<ITaskPosition>) {
		return new TaskPositionModel(data)
	}
}
