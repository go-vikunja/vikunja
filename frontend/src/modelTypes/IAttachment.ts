import type {IAbstract} from './IAbstract'
import type {IFile} from './IFile'
import type {IUser} from './IUser'
import type {IApiErrorResponse} from './IApiError'

export interface IAttachment extends IAbstract {
	id: number
	taskId: number
	createdBy: IUser
	file: IFile
	created: Date
}

export interface IAttachmentUploadResponse {
	success: IAttachment[] | null
	errors: IApiErrorResponse[] | null  
}
