import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ILabel} from './ILabel'

export interface ITaskTemplate extends IAbstract {
	id: number
	title: string
	description: string
	priority: number
	hexColor: string
	percentDone: number
	repeatAfter: number
	repeatMode: number
	labelIds: number[]
	owner: IUser | null
	created: Date
	updated: Date
}
