import {QueryClient} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'
import {projectKeys} from './projects'

const sdk = vi.hoisted(() => ({projectUsersList: vi.fn(), projectUsersCreate: vi.fn(), projectUsersUpdate: vi.fn(), projectUsersDelete: vi.fn(), projectTeamsList: vi.fn(), projectTeamsCreate: vi.fn(), projectTeamsUpdate: vi.fn(), projectTeamsDelete: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))
vi.mock('@/helpers/fetcher', () => ({getApiV2BaseUrl: () => '/api/v2/'}))
vi.mock('@/helpers/auth', () => ({getAuthSessionEpoch: () => 1, getToken: () => null, getTokenIdentity: () => null}))

import {projectShareKeys, projectUserSharesQuery, projectTeamSharesQuery, normalizeSharePermission, createProjectUserShareMutationOptions, updateProjectUserShareMutationOptions, deleteProjectUserShareMutationOptions, createProjectTeamShareMutationOptions, updateProjectTeamShareMutationOptions, deleteProjectTeamShareMutationOptions} from './projectShares'

let client: QueryClient
beforeEach(() => {
	vi.resetAllMocks()
	client = new QueryClient({defaultOptions: {queries: {retry: false}}})
})

describe('project shares', () => {
	it('requests user and team shares for the project', async () => {
		sdk.projectUsersList.mockResolvedValue({data: {items: [{id: 1}], total_pages: 1}})
		sdk.projectTeamsList.mockResolvedValue({data: {items: [{id: 2}], total_pages: 1}})
		expect(await client.fetchQuery(projectUserSharesQuery(7))).toEqual([{id: 1}])
		expect(await client.fetchQuery(projectTeamSharesQuery(7))).toEqual([{id: 2}])
		expect(sdk.projectUsersList).toHaveBeenCalledExactlyOnceWith({path: {project: 7}, query: {page: 1, per_page: 1000}, signal: expect.any(AbortSignal)})
		expect(sdk.projectTeamsList).toHaveBeenCalledExactlyOnceWith({path: {project: 7}, query: {page: 1, per_page: 1000}, signal: expect.any(AbortSignal)})
	})

	it.each([undefined, -1, 3, '2', null])('defaults invalid permission %s to read', permission => {
		expect(normalizeSharePermission(permission)).toBe(0)
	})
	it.each([0, 1, 2])('preserves permission %s', permission => {
		expect(normalizeSharePermission(permission)).toBe(permission)
	})

	it('creates shares and invalidates their list and the project permission snapshot', async () => {
		client.setQueryData(projectShareKeys.users(7), [])
		client.setQueryData(projectShareKeys.teams(7), [])
		client.setQueryData(projectKeys.detail(7), {id: 7, max_permission: 2})
		sdk.projectUsersCreate.mockResolvedValue({data: {id: 90, username: 'sam', permission: 0}})
		await client.getMutationCache().build(client, createProjectUserShareMutationOptions()).execute({projectId: 7, username: 'sam'})
		expect(sdk.projectUsersCreate.mock.calls).toStrictEqual([[{path: {project: 7}, body: {username: 'sam'}}]])
		expect(client.getQueryState(projectShareKeys.users(7))?.isInvalidated).toBe(true)
		expect(client.getQueryState(projectKeys.detail(7))?.isInvalidated).toBe(true)

		client.setQueryData(projectKeys.detail(7), {id: 7, max_permission: 2})
		sdk.projectTeamsCreate.mockResolvedValue({data: {id: 91, team_id: 3, permission: 1}})
		await client.getMutationCache().build(client, createProjectTeamShareMutationOptions()).execute({projectId: 7, teamId: 3})
		expect(sdk.projectTeamsCreate.mock.calls).toStrictEqual([[{path: {project: 7}, body: {team_id: 3}}]])
		expect(client.getQueryState(projectShareKeys.teams(7))?.isInvalidated).toBe(true)
		expect(client.getQueryState(projectKeys.detail(7))?.isInvalidated).toBe(true)
	})

	it('updates user permissions while preserving user ids and neighbouring projects', async () => {
		client.setQueryData(projectShareKeys.users(1), [{id: 7, username: 'sam', permission: 0}])
		client.setQueryData(projectShareKeys.users(2), [{id: 7, username: 'sam', permission: 0}])
		sdk.projectUsersUpdate.mockResolvedValue({data: {id: 99, username: 'sam', permission: 2}})
		await client.getMutationCache().build(client, updateProjectUserShareMutationOptions()).execute({projectId: 1, username: 'sam', permission: 2})
		expect(client.getQueryData(projectShareKeys.users(1))).toEqual([{id: 7, username: 'sam', permission: 2}])
		expect(client.getQueryData(projectShareKeys.users(2))).toEqual([{id: 7, username: 'sam', permission: 0}])
	})

	it('updates team permissions without substituting a relation id', async () => {
		client.setQueryData(projectShareKeys.teams(1), [{id: 7, name: 'ops', permission: 0}])
		sdk.projectTeamsUpdate.mockResolvedValue({data: {id: 99, team_id: 7, permission: 1}})
		await client.getMutationCache().build(client, updateProjectTeamShareMutationOptions()).execute({projectId: 1, teamId: 7, permission: 1})
		expect(client.getQueryData(projectShareKeys.teams(1))).toEqual([{id: 7, name: 'ops', permission: 1}])
	})

	it('removes user and team shares from their own lists', async () => {
		client.setQueryData(projectShareKeys.users(1), [{username: 'sam'}, {username: 'alex'}])
		client.setQueryData(projectShareKeys.teams(1), [{id: 7}, {id: 8}])
		sdk.projectUsersDelete.mockResolvedValue({})
		sdk.projectTeamsDelete.mockResolvedValue({})
		await client.getMutationCache().build(client, deleteProjectUserShareMutationOptions()).execute({projectId: 1, username: 'sam'})
		await client.getMutationCache().build(client, deleteProjectTeamShareMutationOptions()).execute({projectId: 1, teamId: 7})
		expect(client.getQueryData(projectShareKeys.users(1))).toEqual([{username: 'alex'}])
		expect(client.getQueryData(projectShareKeys.teams(1))).toEqual([{id: 8}])
	})
})
