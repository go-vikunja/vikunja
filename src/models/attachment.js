import AbstractModel from './abstractModel'
import UserModel from './user'
import FileModel from './file'

export default class AttachmentModel extends AbstractModel {
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
			file: FileModel,
			createdBy: UserModel,
			created: null,
		}
	}
}
