import type {User as IUser} from '@/client/generated'
import type {IAbstract} from './IAbstract'

export interface IAdminUser extends IUser, IAbstract {
	status: number
	isAdmin: boolean
	issuer: string
	subject?: string
	authProvider?: string
}
