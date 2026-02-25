import type {IAbstract} from '@/modelTypes/IAbstract'

export interface ISession extends IAbstract {
	id: string
	deviceInfo: string
	ipAddress: string
	isCurrent: boolean
	lastActive: Date
	created: Date
}
