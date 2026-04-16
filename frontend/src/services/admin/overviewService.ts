import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {objectToCamelCase} from '@/helpers/case'

export interface AdminOverview {
	users: number
	projects: number
	tasks: number
	shares: {
		linkShares: number
		teamShares: number
		userShares: number
	}
	version: string
	license: {
		enabledProFeatures: string[]
	}
}

export async function getAdminOverview(): Promise<AdminOverview> {
	const {data} = await AuthenticatedHTTPFactory().get('/admin/overview')
	return objectToCamelCase(data) as unknown as AdminOverview
}
