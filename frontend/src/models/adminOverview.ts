import AbstractModel from './abstractModel'
import type {IAdminOverview, IAdminOverviewLicense, IAdminOverviewShares} from '@/modelTypes/IAdminOverview'

export default class AdminOverviewModel extends AbstractModel<IAdminOverview> implements IAdminOverview {
	users = 0
	projects = 0
	tasks = 0
	teams = 0
	shares: IAdminOverviewShares = {
		linkShares: 0,
		teamShares: 0,
		userShares: 0,
	}
	license: IAdminOverviewLicense = {
		licensed: false,
		instanceId: '',
		features: [],
		maxUsers: 0,
		expiresAt: new Date(0),
		validatedAt: new Date(0),
		lastCheckFailed: false,
	}

	constructor(data: Partial<IAdminOverview> = {}) {
		super()
		this.assignData(data)

		this.license.expiresAt = new Date(this.license.expiresAt)
		this.license.validatedAt = new Date(this.license.validatedAt)
	}
}
