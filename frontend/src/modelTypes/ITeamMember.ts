import type {IUser} from './IUser'

export interface ITeamMember extends IUser {
	admin: boolean
	teamId: number
}
