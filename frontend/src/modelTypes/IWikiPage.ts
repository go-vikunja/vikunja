import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'

export interface IWikiPage extends IAbstract {
	id: number
	projectId: number
	parentId: number | null
	title: string
	content: string
	path: string
	isFolder: boolean
	position: number
	createdBy: IUser
	
	created: Date
	updated: Date
	
	// For tree structure in frontend
	children?: IWikiPage[]
}
