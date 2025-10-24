import AbstractModel from '@/models/abstractModel'
import type {IApiToken, ApiTokenLevel} from '@/modelTypes/IApiToken'

export default class ApiTokenModel extends AbstractModel<IApiToken> {
	id = 0
	title = ''
	token = ''
	tokenLevel: ApiTokenLevel = 'standard'
	permissions = null
	expiresAt: Date = null
	created: Date = null
	
	constructor(data: Partial<IApiToken> = {}) {
		super()
		
		this.assignData(data)
		
		this.expiresAt = new Date(this.expiresAt)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
