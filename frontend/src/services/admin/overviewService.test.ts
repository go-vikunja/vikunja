import {describe, it, expect, vi, beforeEach} from 'vitest'

const get = vi.fn()
vi.mock('@/helpers/fetcher', () => ({
	AuthenticatedHTTPFactory: () => ({get}),
}))

import {getAdminOverview} from './overviewService'

describe('overviewService.getAdminOverview', () => {
	beforeEach(() => {
		get.mockReset()
	})

	it('GETs /admin/overview and camel-cases the response', async () => {
		get.mockResolvedValue({data: {
			users: 3,
			projects: 5,
			shares: {link_shares: 1, team_shares: 2, user_shares: 3},
			version: 'v2.3.0',
			license: {enabled_pro_features: ['admin_panel']},
		}})

		const out = await getAdminOverview()

		expect(get).toHaveBeenCalledWith('/admin/overview')
		expect(out.users).toBe(3)
		expect(out.shares.linkShares).toBe(1)
		expect(out.shares.teamShares).toBe(2)
		expect(out.shares.userShares).toBe(3)
		expect(out.license.enabledProFeatures).toEqual(['admin_panel'])
	})

	it('propagates request errors', async () => {
		get.mockRejectedValue(new Error('boom'))
		await expect(getAdminOverview()).rejects.toThrow('boom')
	})
})
