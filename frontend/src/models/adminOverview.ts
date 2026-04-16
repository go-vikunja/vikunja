import AbstractModel from './abstractModel'
import type {IAdminOverview, IAdminOverviewLicense, IAdminOverviewShares} from '@/modelTypes/IAdminOverview'

function parseDate(value: Date | string | null | undefined): Date | null {
	if (!value) {
		return null
	}
	const date = value instanceof Date ? value : new Date(value)
	return isNaN(date.getTime()) ? null : date
}

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
	license: IAdminOverviewLicense = {
		licensed: false,
		instanceId: '',
		features: [],
		maxUsers: 0,
		expiresAt: null,
		validatedAt: null,
		lastCheckFailed: false,
	}

	constructor(data: Partial<IAdminOverview> = {}) {
		super()
		this.assignData(data)

		this.license = {
			...this.license,
			expiresAt: parseDate(this.license?.expiresAt),
			validatedAt: parseDate(this.license?.validatedAt),
		}
	}
}
