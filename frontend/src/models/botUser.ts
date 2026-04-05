import AbstractModel from '@/models/abstractModel'
import type {IBotUser} from '@/modelTypes/IBotUser'

export default class BotUserModel extends AbstractModel<IBotUser> {
	id = 0
	username = ''
	name = ''
	status = 0
	botOwnerId = 0
	created!: Date
	updated!: Date

	constructor(data: Partial<IBotUser> = {}) {
		super()
		this.assignData(data)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
