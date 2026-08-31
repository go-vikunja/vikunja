import type {IAbstract} from '@/modelTypes/IAbstract'

export interface ITaskPosition extends IAbstract {
	position: number
	projectViewId: number
	taskId: number
}
