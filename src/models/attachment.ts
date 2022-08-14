import AbstractModel, { type IAbstract } from './abstractModel'
import UserModel, {type IUser} from './user'
import FileModel, {type IFile} from './file'

export interface IAttachment extends IAbstract {
	id: number
	taskId: number
	createdBy: IUser
	file: IFile
	created: Date
}

export default class AttachmentModel extends AbstractModel implements IAttachment {
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
