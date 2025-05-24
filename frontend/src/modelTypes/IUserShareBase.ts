import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {Right} from '@/constants/rights'

export interface IUserShareBase extends IAbstract {
	userId: IUser['id']
	right: Right

	created: Date
	updated: Date
}
