import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {Permission} from '@/constants/permissions'

export interface IUserShareBase extends IAbstract {
	username: IUser['username']
	permission: Permission

	created: Date
	updated: Date
}
