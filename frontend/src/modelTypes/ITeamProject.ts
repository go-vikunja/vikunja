import type {ITeamShareBase} from './ITeamShareBase'
import type {IProject} from './IProject'

export interface ITeamProject extends ITeamShareBase {
	projectId: IProject['id']
}
