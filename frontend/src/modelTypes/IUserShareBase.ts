import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {Right} from '@/constants/rights'

export interface IUserShareBase extends IAbstract {
	username: IUser['username']
	right: Right

	created: Date
	updated: Date
}
