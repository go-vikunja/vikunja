import type {IUser} from './IUser'
import type {IList} from './IList'

export interface ITeamMember extends IUser {
	admin: boolean
	teamId: IList['id']
}
