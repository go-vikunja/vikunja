import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type { Permission } from '@/constants/permissions'

export interface ILinkShare extends IAbstract {
	id: number
	hash: string
	permission: Permission
	sharedBy: IUser
	sharingType: number // FIXME: use correct numbers
	projectId: number
	name: string
	password: string

	created: Date
	updated: Date
}
