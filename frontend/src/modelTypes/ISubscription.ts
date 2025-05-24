import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'

export interface ISubscription extends IAbstract {
	id: number
	entity: string // FIXME: correct type?
	entityId: number // FIXME: correct type?
	user: IUser

	created: Date
}
