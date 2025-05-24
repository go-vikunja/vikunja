import type {IAbstract} from './IAbstract'

export interface IPasswordReset extends IAbstract {
	token: string
	newPassword: string
	email: string
}
