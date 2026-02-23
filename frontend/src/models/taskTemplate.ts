import AbstractModel from './abstractModel'
import UserModel from './user'

import type {ITaskTemplate} from '@/modelTypes/ITaskTemplate'
import type {IUser} from '@/modelTypes/IUser'

export default class TaskTemplateModel extends AbstractModel<ITaskTemplate> implements ITaskTemplate {
	id = 0
	title = ''
	description = ''
	priority = 0
	hexColor = ''
	percentDone = 0
	repeatAfter = 0
	repeatMode = 0
	labelIds: number[] = []
	owner: IUser | null = null
	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<ITaskTemplate> = {}) {
		super()
		this.assignData(data)

		this.owner = this.owner ? new UserModel(this.owner) : null
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
