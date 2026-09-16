import {QueryClient} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'
import type {TeamReadBody} from '@/client/generated'
import {success} from '@/message'

const sdk = vi.hoisted(() => ({teamsList: vi.fn(), teamsRead: vi.fn(), teamsCreate: vi.fn(), teamsUpdate: vi.fn(), teamsDelete: vi.fn(), teamsMembersAdd: vi.fn(), teamsMembersRemove: vi.fn(), teamsMembersToggleAdmin: vi.fn()}))
const session = vi.hoisted(() => ({epoch: 1}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))
vi.mock('@/helpers/auth', () => ({getAuthSessionEpoch: () => session.epoch, getToken: () => null, getTokenIdentity: () => null}))
vi.mock('@/helpers/fetcher', () => ({getApiV2BaseUrl: () => '/api/v2/'}))

import {teamKeys, teamsQuery, updateTeamMutationOptions, deleteTeamMutationOptions, addTeamMemberMutationOptions, removeTeamMemberMutationOptions, leaveTeamMutationOptions, toggleTeamMemberAdminMutationOptions} from './teams'

let client: QueryClient
beforeEach(() => {
	vi.resetAllMocks()
	session.epoch = 1
	client = new QueryClient({defaultOptions: {queries: {retry: false}}})
})

describe('teams', () => {
	it('passes list filters to the page request', async () => {
		sdk.teamsList.mockResolvedValue({data: {items: [{id: 1}], total_pages: 1}})
		expect(await client.fetchQuery(teamsQuery('ops', true))).toEqual([{id: 1}])
		expect(sdk.teamsList).toHaveBeenCalledWith({query: {q: 'ops', include_public: true, page: 1, per_page: 1000}, signal: expect.any(AbortSignal)})
	})

	it('updates mounted team caches without dropping read permissions', async () => {
		client.setQueryData(teamKeys.list(), [{id: 1, name: 'old'}])
		client.setQueryData(teamKeys.detail(1), {id: 1, name: 'old', max_permission: 2, members: []})
		sdk.teamsUpdate.mockResolvedValue({data: {id: 1, name: 'new'}})
		await client.getMutationCache().build(client, updateTeamMutationOptions()).execute({id: 1, team: {name: 'new'}})
		expect(client.getQueryData(teamKeys.list())).toEqual([{id: 1, name: 'new'}])
		expect(client.getQueryData(teamKeys.detail(1))).toMatchObject({name: 'new', max_permission: 2})
		expect(client.getQueryState(teamKeys.list())?.isInvalidated).toBe(true)
	})

	it('removes a deleted team from existing lists and drops its detail', async () => {
		client.setQueryData(teamKeys.list(), [{id: 1}, {id: 2}])
		client.setQueryData(teamKeys.detail(1), {id: 1})
		sdk.teamsDelete.mockResolvedValue({})
		await client.getMutationCache().build(client, deleteTeamMutationOptions()).execute(1)
		expect(client.getQueryData(teamKeys.list())).toEqual([{id: 2}])
		expect(client.getQueryData(teamKeys.detail(1))).toBeUndefined()
	})

	it('reconciles member relations by username, not the relation id', async () => {
		client.setQueryData(teamKeys.detail(1), {id: 1, members: [{id: 7, username: 'sam', admin: false}]})
		sdk.teamsMembersToggleAdmin.mockResolvedValue({data: {id: 99, username: 'sam', admin: true}})
		await client.getMutationCache().build(client, toggleTeamMemberAdminMutationOptions()).execute({teamId: 1, username: 'sam'})
		expect(client.getQueryData<TeamReadBody>(teamKeys.detail(1))?.members).toEqual([{id: 7, username: 'sam', admin: true}])
		sdk.teamsMembersRemove.mockResolvedValue({})
		await client.getMutationCache().build(client, removeTeamMemberMutationOptions()).execute({teamId: 1, username: 'sam'})
		expect(client.getQueryData<TeamReadBody>(teamKeys.detail(1))?.members).toEqual([])
	})

	it('drops the left team detail instead of refetching it', async () => {
		client.setQueryData(teamKeys.list(), [{id: 1}])
		client.setQueryData(teamKeys.detail(1), {id: 1, members: [{username: 'sam'}]})
		sdk.teamsMembersRemove.mockResolvedValue({})
		await client.getMutationCache().build(client, leaveTeamMutationOptions()).execute({teamId: 1, username: 'sam'})
		expect(sdk.teamsMembersRemove).toHaveBeenCalledWith({path: {team: 1, user: 'sam'}})
		expect(client.getQueryState(teamKeys.detail(1))).toBeUndefined()
		expect(client.getQueryState(teamKeys.list())?.isInvalidated).toBe(true)
		expect(sdk.teamsRead).not.toHaveBeenCalled()
	})

	it('invalidates the parent after adding a member without inventing a user', async () => {
		client.setQueryData(teamKeys.detail(1), {id: 1, members: []})
		sdk.teamsMembersAdd.mockResolvedValue({data: {id: 99, username: 'sam'}})
		await client.getMutationCache().build(client, addTeamMemberMutationOptions()).execute({teamId: 1, username: 'sam'})
		expect(client.getQueryData<TeamReadBody>(teamKeys.detail(1))?.members).toEqual([])
		expect(client.getQueryState(teamKeys.detail(1))?.isInvalidated).toBe(true)
	})

	it('keeps failed membership changes out of the cache', async () => {
		client.setQueryData(teamKeys.detail(1), {id: 1, members: [{username: 'sam', admin: false}]})
		sdk.teamsMembersToggleAdmin.mockRejectedValue(new Error('denied'))
		await expect(client.getMutationCache().build(client, toggleTeamMemberAdminMutationOptions()).execute({teamId: 1, username: 'sam'})).rejects.toThrow('denied')
		expect(client.getQueryData<TeamReadBody>(teamKeys.detail(1))?.members?.[0].admin).toBe(false)
		expect(success).not.toHaveBeenCalled()
	})

	it('fences cache writes and notifications after an identity change', async () => {
		client.setQueryData(teamKeys.detail(1), {id: 1, name: 'other identity'})
		sdk.teamsUpdate.mockImplementation(async () => { session.epoch++; return {data: {id: 1, name: 'old identity'}} })
		await expect(client.getMutationCache().build(client, updateTeamMutationOptions()).execute({id: 1, team: {name: 'old identity'}})).rejects.toThrow('Client request context changed')
		expect(client.getQueryData(teamKeys.detail(1))).toEqual({id: 1, name: 'other identity'})
		expect(success).not.toHaveBeenCalled()
	})
})
