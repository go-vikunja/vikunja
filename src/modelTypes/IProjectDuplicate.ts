import type {IAbstract} from './IAbstract'
import type {IProject} from './IProject'
import type {INamespace} from './INamespace'

export interface IProjectDuplicate extends IAbstract {
	projectId: number
	namespaceId: INamespace['id']
	project: IProject
}