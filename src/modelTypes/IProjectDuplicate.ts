import type {IAbstract} from './IAbstract'
import type {IProject} from './IProject'

export interface IProjectDuplicate extends IAbstract {
	projectId: number
	project: IProject
	parentProjectId: number
}