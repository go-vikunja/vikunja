import AbstractModel from './abstractModel'
import UserModel from './user'
import FileModel from './file'
import type { IUser } from '@/modelTypes/IUser'
import type { IFile } from '@/modelTypes/IFile'
import type { IAttachment } from '@/modelTypes/IAttachment'

export const SUPPORTED_IMAGE_SUFFIX = ['.jpg', '.png', '.bmp', '.gif']

export default class AttachmentModel extends AbstractModel<IAttachment> implements IAttachment {
	id = 0
	taskId = 0
	createdBy: IUser = UserModel
	file: IFile = FileModel
	created: Date = null

	constructor(data: Partial<IAttachment>) {
		super()
		this.assignData(data)

		this.createdBy = new UserModel(this.createdBy)
		this.file = new FileModel(this.file)
		this.created = new Date(this.created)
	}
}
