import type {IProjectView} from '@/modelTypes/IProjectView'
import type {IAbstract} from '@/modelTypes/IAbstract'

export interface ITaskPosition extends IAbstract {
	position: number
	projectViewId: IProjectView['id']
	taskId: number
}
