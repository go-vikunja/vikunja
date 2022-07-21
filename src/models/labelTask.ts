import AbstractModel from './abstractModel'

interface ILabel extends AbstractModel {
	id: number
	taskId: number
	labelId: number
}

export default class LabelTask extends AbstractModel implements ILabel {
	declare id: number
	declare taskId: number
	declare labelId: number

	defaults() {
		return {
			id: 0,
			taskId: 0,
			labelId: 0,
		}
	}
}