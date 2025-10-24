import type {IAbstract} from '@/modelTypes/IAbstract'

export interface IApiPermission {
	[key: string]: string[]
}

export type ApiTokenLevel = 'standard' | 'admin'

export interface IApiToken extends IAbstract {
	id: number
	title: string
	token: string
	tokenLevel: ApiTokenLevel
	permissions: IApiPermission
	expiresAt: Date
	created: Date
}
