import type {IAbstract} from './IAbstract'

export interface IAdminOverviewShares {
	linkShares: number
	teamShares: number
	userShares: number
}

export interface IAdminOverviewLicense {
	licensed: boolean
	instanceId: string
	features: string[]
	maxUsers: number
	expiresAt: Date
	validatedAt: Date
	lastCheckFailed: boolean
}

export interface IAdminOverview extends IAbstract {
	users: number
	projects: number
	tasks: number
	teams: number
	shares: IAdminOverviewShares
	version: string
	license: IAdminOverviewLicense
}
