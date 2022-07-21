import AbstractModel from './abstractModel'
import UserModel, {type IUser} from './user'
import FileModel, {type IFile} from './file'

export interface IAttachment extends AbstractModel {
	id: number
	taskId: number
	createdBy: IUser
	file: IFile
	created: Date
}

export default class AttachmentModel extends AbstractModel implements IAttachment {
	declare id: number
	declare taskId: number
	createdBy: IUser
	file: IFile
	created: Date

	constructor(data) {
		super(data)
		this.createdBy = new UserModel(this.createdBy)
		this.file = new FileModel(this.file)
		this.created = new Date(this.created)
	}

	defaults() {
		return {
			id: 0,
			taskId: 0,
			createdBy: UserModel,
			file: FileModel,
			created: null,
		}
	}
}
