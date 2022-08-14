import AbstractModel, { type IAbstract } from './abstractModel'

export interface ILabelTask extends IAbstract {
	id: number
	taskId: number
	labelId: number
}

export default class LabelTask extends AbstractModel implements ILabelTask {
	id = 0
	taskId = 0
	labelId = 0

	constructor(data: Partial<ILabelTask>) {
		super()
		this.assignData(data)
	}
}