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
			tasks: 42,
			shares: {link_shares: 1, team_shares: 2, user_shares: 3},
			version: 'v2.3.0',
			license: {
				licensed: true,
				instance_id: 'inst-1',
				features: ['admin_panel'],
				max_users: 0,
				expires_at: '2099-01-01T00:00:00Z',
				validated_at: '2026-04-01T00:00:00Z',
				last_check_failed: false,
			},
		}})

		const out = await getAdminOverview()

		expect(get).toHaveBeenCalledWith('/admin/overview')
		expect(out.users).toBe(3)
		expect(out.tasks).toBe(42)
		expect(out.shares.linkShares).toBe(1)
		expect(out.shares.teamShares).toBe(2)
		expect(out.shares.userShares).toBe(3)
		expect(out.license.licensed).toBe(true)
		expect(out.license.features).toEqual(['admin_panel'])
		expect(out.license.instanceId).toBe('inst-1')
	})

	it('propagates request errors', async () => {
		get.mockRejectedValue(new Error('boom'))
		await expect(getAdminOverview()).rejects.toThrow('boom')
	})
})
