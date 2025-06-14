import AbstractModel from './abstractModel'

import type { ILabelTask } from '@/modelTypes/ILabelTask'

export default class LabelTask extends AbstractModel<ILabelTask> implements ILabelTask {
	id = 0
	taskId = 0
	labelId = 0

	constructor(data: Partial<ILabelTask>) {
		super()
		this.assignData(data)
	}
}
