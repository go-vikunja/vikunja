import type {IAbstract} from './IAbstract'

export interface ILabelTask extends IAbstract {
	id: number
	taskId: number
	labelId: number
}
