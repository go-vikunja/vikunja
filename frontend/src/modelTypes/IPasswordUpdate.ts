import type {IAbstract} from './IAbstract'

export interface IPasswordUpdate extends IAbstract {
	newPassword: string
	oldPassword: string
}
