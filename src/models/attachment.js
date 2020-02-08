import AbstractModel from './abstractModel'
import UserModel from './user'
import FileModel from './file'

export default class AttachmentModel extends AbstractModel {
	constructor(data) {
		super(data)
		this.created_by = new UserModel(this.created_by)
		this.file = new FileModel(this.file)
		this.created = new Date(this.created)
	}

	defaults() {
		return {
			id: 0,
			task_id: 0,
			file: FileModel,
			created_by: UserModel,
			created: null,
		}
	}
}
