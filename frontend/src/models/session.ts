import AbstractModel from '@/models/abstractModel'
import type {ISession} from '@/modelTypes/ISession'

export default class SessionModel extends AbstractModel<ISession> implements ISession {
	id = ''
	deviceInfo = ''
	ipAddress = ''
	isCurrent = false
	lastActive: Date = null
	created: Date = null

	constructor(data: Partial<ISession> = {}) {
		super()

		this.assignData(data)

		if (this.lastActive) {
			this.lastActive = new Date(this.lastActive)
		}
		if (this.created) {
			this.created = new Date(this.created)
		}
	}
}
