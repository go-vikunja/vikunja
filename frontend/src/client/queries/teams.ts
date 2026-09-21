import {queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'
import {teamsList, teamsRead, teamsCreate, teamsUpdate, teamsDelete, teamsMembersAdd, teamsMembersRemove, teamsMembersToggleAdmin} from '@/client/generated'
import type {Team, TeamReadBody, TeamWritable, TeamMemberWritable} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {fetchAllPages} from './fetchAllPages'
import {API_MAX_PER_PAGE} from './pagination'
import {projectKeys} from './projects'
import {userSearchKeys} from './userSearch'
import {i18n} from '@/i18n'

export const teamKeys = {
	lists: ['teams', 'list'] as const,
	list: (search = '', includePublic = false) => [...teamKeys.lists, search, includePublic] as const,
	detail: (id: number) => ['teams', 'detail', id] as const,
}

export function createTeamDraft(team: TeamWritable = {}): Required<TeamWritable> {
	return {name: team.name ?? '', description: team.description ?? '', is_public: team.is_public ?? false}
}

export function teamsQuery(search = '', includePublic = false) {
	return queryOptions({
		queryKey: teamKeys.list(search, includePublic),
		queryFn: ({signal}) => fetchAllPages(async page => (await teamsList({query: {q: search, include_public: includePublic, page, per_page: API_MAX_PER_PAGE}, signal})).data),
	})
}

export function teamQuery(id: number) {
	return queryOptions({
		queryKey: teamKeys.detail(id),
		queryFn: async ({signal}) => (await teamsRead({path: {id}, signal})).data,
	})
}

function updateCachedTeam(client: QueryClient, id: number, update: (team: TeamReadBody) => TeamReadBody) {
	client.setQueryData<TeamReadBody>(teamKeys.detail(id), current => current ? update(current) : current)
	client.setQueriesData<Team[]>({queryKey: teamKeys.lists}, current => current?.map(team => team.id === id ? update(team) : team))
}

async function invalidateTeams(client: QueryClient, id?: number) {
	await Promise.all([
		client.invalidateQueries({queryKey: teamKeys.lists}),
		...(id ? [client.invalidateQueries({queryKey: teamKeys.detail(id)})] : []),
	])
}

export function createTeamMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (team: TeamWritable) => (await teamsCreate({body: team})).data,
		onSettled: (_input, client) => invalidateTeams(client),
		successMessage: () => i18n.global.t('team.create.success'),
	})
}

export function updateTeamMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, team}: {id: number, team: TeamWritable}) => (await teamsUpdate({path: {id}, body: team})).data,
		onSuccess: (updated, {id}, client) => updateCachedTeam(client, id, team => ({...team, ...updated})),
		onSettled: ({id}, client) => invalidateTeams(client, id),
		successMessage: () => i18n.global.t('team.edit.success'),
	})
}

export function deleteTeamMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => { await teamsDelete({path: {id}}) },
		onSuccess: (_data, id, client) => {
			client.setQueriesData<Team[]>({queryKey: teamKeys.lists}, current => current?.filter(team => team.id !== id))
			client.removeQueries({queryKey: teamKeys.detail(id), exact: true})
		},
		onSettled: (_id, client) => invalidateTeams(client),
		successMessage: () => i18n.global.t('team.edit.delete.success'),
	})
}

type MemberInput = {teamId: number} & Required<Pick<TeamMemberWritable, 'username'>>

async function invalidateMembership(client: QueryClient, teamId?: number) {
	await Promise.all([
		invalidateTeams(client, teamId),
		client.invalidateQueries({queryKey: userSearchKeys.all, refetchType: 'none'}),
		client.invalidateQueries({queryKey: projectKeys.all, refetchType: 'none'}),
	])
}

export function addTeamMemberMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({teamId, username}: MemberInput) => (await teamsMembersAdd({path: {team: teamId}, body: {username, admin: false}})).data,
		onSettled: ({teamId}, client) => invalidateMembership(client, teamId),
		successMessage: () => i18n.global.t('team.edit.userAddedSuccess'),
	})
}

async function removeTeamMember({teamId, username}: MemberInput) {
	await teamsMembersRemove({path: {team: teamId, user: username}})
}

export function removeTeamMemberMutationOptions() {
	return contextMutationOptions({
		mutationFn: removeTeamMember,
		onSuccess: (_data, {teamId, username}, client) => updateCachedTeam(client, teamId, team => ({...team, members: team.members?.filter(member => member.username !== username)})),
		onSettled: ({teamId}, client) => invalidateMembership(client, teamId),
		successMessage: () => i18n.global.t('team.edit.deleteUser.success'),
	})
}

export function leaveTeamMutationOptions() {
	return contextMutationOptions({
		mutationFn: removeTeamMember,
		onSuccess: (_data, {teamId}, client) => client.removeQueries({queryKey: teamKeys.detail(teamId), exact: true}),
		onSettled: (_input, client) => invalidateMembership(client),
		successMessage: () => i18n.global.t('team.edit.leave.success'),
	})
}

export function toggleTeamMemberAdminMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({teamId, username}: MemberInput) => (await teamsMembersToggleAdmin({path: {team: teamId, user: username}})).data,
		onSuccess: (updated, {teamId, username}, client) => updateCachedTeam(client, teamId, team => ({...team, members: team.members?.map(member => member.username === username ? {...member, admin: updated.admin} : member)})),
		onSettled: ({teamId}, client) => invalidateMembership(client, teamId),
		successMessage: (updated) => i18n.global.t(updated.admin ? 'team.edit.madeAdmin' : 'team.edit.madeMember'),
	})
}

export const useCreateTeamMutation = () => useMutation(createTeamMutationOptions())
export const useUpdateTeamMutation = () => useMutation(updateTeamMutationOptions())
export const useDeleteTeamMutation = () => useMutation(deleteTeamMutationOptions())
export const useAddTeamMemberMutation = () => useMutation(addTeamMemberMutationOptions())
export const useRemoveTeamMemberMutation = () => useMutation(removeTeamMemberMutationOptions())
export const useLeaveTeamMutation = () => useMutation(leaveTeamMutationOptions())
export const useToggleTeamMemberAdminMutation = () => useMutation(toggleTeamMemberAdminMutationOptions())
