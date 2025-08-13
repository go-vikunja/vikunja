import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITeamMember} from './ITeamMember'
import type {Permission} from '@/constants/permissions'

export interface ITeam extends IAbstract {
	id: number
	name: string
	description: string
	members: ITeamMember[]
	permission: Permission
	externalId: string
	isPublic: boolean

	createdBy: IUser
	created: Date
	updated: Date
}
