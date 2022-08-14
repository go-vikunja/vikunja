import AbstractModel, { type IAbstract } from './abstractModel'

export interface ILabelTask extends IAbstract {
	id: number
	taskId: number
	labelId: number
}

export default class LabelTask extends AbstractModel implements ILabelTask {
	id!: number
	taskId!: number
	labelId!: number

	defaults() {
		return {
			id: 0,
			taskId: 0,
			labelId: 0,
		}
	}
}