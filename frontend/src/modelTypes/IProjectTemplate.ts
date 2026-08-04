import type {IAbstract} from './IAbstract'
import type {IProject} from './IProject'

export interface IProjectTemplate extends IAbstract {
	projectId: number
	project: IProject | null
}
