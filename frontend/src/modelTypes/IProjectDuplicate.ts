import type {IAbstract} from './IAbstract'
import type {IProject} from './IProject'

export interface IProjectDuplicate extends IAbstract {
	projectId: number
	duplicatedProject: IProject | null
	parentProjectId: IProject['id']
}
