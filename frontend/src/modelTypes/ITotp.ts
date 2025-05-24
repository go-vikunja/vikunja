import type {IAbstract} from './IAbstract'

export interface ITotp extends IAbstract {
	secret: string
	enabled: boolean
	url: string
}
