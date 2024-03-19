import AbstractModel from '@/models/abstractModel'
import type {ITaskPosition} from '@/modelTypes/ITaskPosition'

export default class TaskPositionModel extends AbstractModel<ITaskPosition> implements ITaskPosition {
	position = 0
	projectViewId = 0
	taskId = 0

	constructor(data: Partial<ITaskPosition>) {
		super()
		this.assignData(data)
	}
}
