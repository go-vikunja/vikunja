import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type { Right } from '@/constants/rights'

export interface ILinkShare extends IAbstract {
	id: number
	hash: string
	right: Right
	sharedBy: IUser
	sharingType: number // FIXME: use correct numbers
	projectId: number
	name: string
	password: string

	created: Date
	updated: Date
}
