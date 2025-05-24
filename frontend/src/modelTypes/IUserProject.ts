import type {IUserShareBase} from './IUserShareBase'
import type {IProject} from './IProject'

export interface IUserProject extends IUserShareBase {
	projectId: IProject['id']
}
