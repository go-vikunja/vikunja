import type {IUser} from './IUser'
import type {IProject} from './IProject'

export interface ITeamMember extends IUser {
	admin: boolean
	teamId: IProject['id']
}
