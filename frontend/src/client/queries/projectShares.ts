import {queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'
import {projectUsersList, projectUsersCreate, projectUsersUpdate, projectUsersDelete, projectTeamsList, projectTeamsCreate, projectTeamsUpdate, projectTeamsDelete} from '@/client/generated'
import type {UserWithPermission, TeamWithPermission} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {fetchAllPages} from './fetchAllPages'
import {API_MAX_PER_PAGE} from './pagination'
import {projectKeys} from './projects'
import {userSearchKeys} from './userSearch'
import {PERMISSIONS, type Permission} from '@/constants/permissions'
import {i18n} from '@/i18n'

export const projectShareKeys = {
	users: (projectId: number) => ['project-shares', projectId, 'users'] as const,
	teams: (projectId: number) => ['project-shares', projectId, 'teams'] as const,
}

export function normalizeSharePermission(permission: unknown): Permission {
	return permission === PERMISSIONS.ADMIN || permission === PERMISSIONS.READ_WRITE ? permission : PERMISSIONS.READ
}

async function invalidateShares(client: QueryClient, projectId: number, kind: 'users' | 'teams') {
	await Promise.all([
		client.invalidateQueries({queryKey: projectShareKeys[kind](projectId)}),
		client.invalidateQueries({queryKey: projectKeys.detail(projectId)}),
		client.invalidateQueries({queryKey: projectKeys.list(), refetchType: 'none'}),
		client.invalidateQueries({queryKey: userSearchKeys.all, refetchType: 'none'}),
	])
}

function shareSuccess(kind: 'users' | 'teams', action: 'addedSuccess' | 'updatedSuccess' | 'removeSuccess') {
	return i18n.global.t(`project.share.userTeam.${action}`, {
		type: i18n.global.t(kind === 'users' ? 'project.share.userTeam.typeUser' : 'project.share.userTeam.typeTeam', 1),
		sharable: i18n.global.t('project.list.title'),
	})
}

type UserShareInput = {projectId: number, username: string}
type TeamShareInput = {projectId: number, teamId: number}

export function projectUserSharesQuery(projectId: number) {
	return queryOptions({
		queryKey: projectShareKeys.users(projectId),
		queryFn: ({signal}) => fetchAllPages(async page => (await projectUsersList({path: {project: projectId}, query: {page, per_page: API_MAX_PER_PAGE}, signal})).data),
	})
}

export function createProjectUserShareMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({projectId, username}: UserShareInput) => (await projectUsersCreate({path: {project: projectId}, body: {username}})).data,
		onSettled: ({projectId}, client) => invalidateShares(client, projectId, 'users'),
		successMessage: () => shareSuccess('users', 'addedSuccess'),
	})
}
export const useCreateProjectUserShareMutation = () => useMutation(createProjectUserShareMutationOptions())

export function updateProjectUserShareMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({projectId, username, permission}: UserShareInput & {permission: Permission}) => (await projectUsersUpdate({path: {project: projectId, user: username}, body: {username, permission}})).data,
		onSuccess: (updated, {projectId, username}, client) => {
			client.setQueryData<UserWithPermission[]>(projectShareKeys.users(projectId), current => current?.map(item => item.username === username ? {...item, permission: updated.permission} : item))
		},
		onSettled: ({projectId}, client) => invalidateShares(client, projectId, 'users'),
		successMessage: () => shareSuccess('users', 'updatedSuccess'),
	})
}
export const useUpdateProjectUserShareMutation = () => useMutation(updateProjectUserShareMutationOptions())

export function deleteProjectUserShareMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({projectId, username}: UserShareInput) => { await projectUsersDelete({path: {project: projectId, user: username}}) },
		onSuccess: (_data, {projectId, username}, client) => {
			client.setQueryData<UserWithPermission[]>(projectShareKeys.users(projectId), current => current?.filter(item => item.username !== username))
		},
		onSettled: ({projectId}, client) => invalidateShares(client, projectId, 'users'),
		successMessage: () => shareSuccess('users', 'removeSuccess'),
	})
}
export const useDeleteProjectUserShareMutation = () => useMutation(deleteProjectUserShareMutationOptions())

export function projectTeamSharesQuery(projectId: number) {
	return queryOptions({
		queryKey: projectShareKeys.teams(projectId),
		queryFn: ({signal}) => fetchAllPages(async page => (await projectTeamsList({path: {project: projectId}, query: {page, per_page: API_MAX_PER_PAGE}, signal})).data),
	})
}

export function createProjectTeamShareMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({projectId, teamId}: TeamShareInput) => (await projectTeamsCreate({path: {project: projectId}, body: {team_id: teamId}})).data,
		onSettled: ({projectId}, client) => invalidateShares(client, projectId, 'teams'),
		successMessage: () => shareSuccess('teams', 'addedSuccess'),
	})
}
export const useCreateProjectTeamShareMutation = () => useMutation(createProjectTeamShareMutationOptions())

export function updateProjectTeamShareMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({projectId, teamId, permission}: TeamShareInput & {permission: Permission}) => (await projectTeamsUpdate({path: {project: projectId, team: teamId}, body: {team_id: teamId, permission}})).data,
		onSuccess: (updated, {projectId, teamId}, client) => {
			client.setQueryData<TeamWithPermission[]>(projectShareKeys.teams(projectId), current => current?.map(item => item.id === teamId ? {...item, permission: updated.permission} : item))
		},
		onSettled: ({projectId}, client) => invalidateShares(client, projectId, 'teams'),
		successMessage: () => shareSuccess('teams', 'updatedSuccess'),
	})
}
export const useUpdateProjectTeamShareMutation = () => useMutation(updateProjectTeamShareMutationOptions())

export function deleteProjectTeamShareMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({projectId, teamId}: TeamShareInput) => { await projectTeamsDelete({path: {project: projectId, team: teamId}}) },
		onSuccess: (_data, {projectId, teamId}, client) => {
			client.setQueryData<TeamWithPermission[]>(projectShareKeys.teams(projectId), current => current?.filter(item => item.id !== teamId))
		},
		onSettled: ({projectId}, client) => invalidateShares(client, projectId, 'teams'),
		successMessage: () => shareSuccess('teams', 'removeSuccess'),
	})
}
export const useDeleteProjectTeamShareMutation = () => useMutation(deleteProjectTeamShareMutationOptions())
