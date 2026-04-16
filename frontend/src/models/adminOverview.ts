import AbstractModel from './abstractModel'
import type {IAdminOverview, IAdminOverviewLicense, IAdminOverviewShares} from '@/modelTypes/IAdminOverview'

export default class AdminOverviewModel extends AbstractModel<IAdminOverview> implements IAdminOverview {
	users = 0
	projects = 0
	tasks = 0
	shares: IAdminOverviewShares = {
		linkShares: 0,
		teamShares: 0,
		userShares: 0,
	}
	version = ''
	declare license: IAdminOverviewLicense

	constructor(data: Partial<IAdminOverview> = {}) {
		super()
		this.assignData(data)

		this.license.expiresAt = new Date(this.license.expiresAt)
		this.license.validatedAt = new Date(this.license.validatedAt)
	}
}
