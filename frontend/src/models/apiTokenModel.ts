import AbstractModel from '@/models/abstractModel'
import type {IApiToken, IApiPermission} from '@/modelTypes/IApiToken'

export default class ApiTokenModel extends AbstractModel<IApiToken> {
	id = 0
	title = ''
	token = ''
	permissions: IApiPermission = {}
	expiresAt: Date = new Date(0)
	created: Date = new Date(0)
	ownerId = 0
	
	constructor(data: Partial<IApiToken> = {}) {
		super()
		
		this.assignData(data)
		
		this.expiresAt = new Date(this.expiresAt)
		this.created = new Date(this.created)
	}
}
