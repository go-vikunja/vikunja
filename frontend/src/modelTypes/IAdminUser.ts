import type {IUser} from './IUser'

export interface IAdminUser extends IUser {
	status: number
	isAdmin: boolean
	issuer: string
	subject?: string
	authProvider?: string
}
