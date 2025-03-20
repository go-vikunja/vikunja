import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITeamMember} from './ITeamMember'
import type {Right} from '@/constants/rights'

export interface ITeam extends IAbstract {
	id: number
	name: string
	description: string
	members: ITeamMember[]
	right: Right
	externalId: string
	isPublic: boolean

	createdBy: IUser
	created: Date
	updated: Date
}