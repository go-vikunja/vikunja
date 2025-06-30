import AbstractModel from '@/models/abstractModel'
import type {IApiToken} from '@/modelTypes/IApiToken'

export default class ApiTokenModel extends AbstractModel<IApiToken> {
	id = 0
	title = ''
	token = ''
	permissions = null
	expiresAt: Date | null = null
	created: Date | null = null
	updated: Date | null = null
	
	constructor(data: Partial<IApiToken> = {}) {
		super()
		
		this.assignData(data)
		
		if (this.expiresAt) this.expiresAt = new Date(this.expiresAt)
		if (this.created) this.created = new Date(this.created)
		if (this.updated) this.updated = new Date(this.updated)
	}
}
