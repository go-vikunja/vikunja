import AbstractModel from './abstractModel'

export default class LabelTask extends AbstractModel {
	id: number
	taskId: number
	labelId: number

	defaults() {
		return {
			id: 0,
			taskId: 0,
			labelId: 0,
		}
	}
}