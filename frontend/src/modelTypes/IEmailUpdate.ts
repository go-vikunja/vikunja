import type {IAbstract} from './IAbstract'

export interface IEmailUpdate extends IAbstract {
	newEmail: string
	password: string
}
