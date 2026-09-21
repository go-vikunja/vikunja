import AbstractModel from './abstractModel'
import type {IAdminUser} from '@/modelTypes/IAdminUser'

export default class AdminUserModel extends AbstractModel<IAdminUser> implements IAdminUser {
	declare status: number
	declare isAdmin: boolean
	declare issuer: string
	declare subject?: string
	declare authProvider?: string

	constructor(data: Partial<IAdminUser> = {}) {
		super()
		this.assignData(data)
	}
}
