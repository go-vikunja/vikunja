import type {IAbstract} from './IAbstract'
import type {IFile} from './IFile'
import type {IUser} from './IUser'

export interface IAttachment extends IAbstract {
	id: number
	taskId: number
	createdBy: IUser
	file: IFile
	created: Date
}
