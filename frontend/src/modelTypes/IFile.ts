import type {IAbstract} from './IAbstract'

export interface IFile extends IAbstract {
	id: number
	mime: string
	name: string
	size: number

	created: Date
} 
