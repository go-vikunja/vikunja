import {queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'
import {
	adminOverview, adminUsersList, adminUsersCreate, adminUsersPatchAdmin, adminUsersPatchStatus,
	adminUsersSetPassword, adminUsersPasswordResetEmail, adminUsersDelete,
	adminProjectsList, adminProjectsPatchOwner, adminInviteLinksList, adminInviteLinksCreate,
	adminInviteLinksDelete, adminTeamsList,
	type AdminUser, type PaginatedAdminUser, type CreateUserBodyWritable, type User,
	type AdminInviteLinksCreateData,
} from '@/client/generated'
import {captureClientRequestContext, assertClientRequestContext} from '@/client/requestContext'
import {contextMutationOptions} from './contextMutation'
import {fetchAllPages} from './fetchAllPages'
import {projectKeys} from './projects'
import {accountKeys} from './account'
import {i18n} from '@/i18n'

export type DeleteUserMode = 'now' | 'scheduled'
export const adminKeys = {
	all: ['admin'] as const,
	overview: ['admin', 'overview'] as const,
	users: ['admin', 'users'] as const,
	usersPage: (q: string, page: number) => ['admin', 'users', q, page] as const,
	userSearch: (q: string) => ['admin', 'userSearch', q] as const,
	projects: ['admin', 'projects'] as const,
	projectsPage: (page: number) => ['admin', 'projects', page] as const,
	invites: ['admin', 'invites'] as const,
	invitesPage: (page: number) => ['admin', 'invites', page] as const,
	teams: (q: string) => ['admin', 'teams', q] as const,
}
export function adminOverviewQuery() {
	return queryOptions({queryKey: adminKeys.overview, queryFn: async ({signal}) => (await adminOverview({signal})).data})
}
export function adminUsersQuery(q = '', page = 1) {
	return queryOptions({queryKey: adminKeys.usersPage(q, page), queryFn: async ({signal}) => (await adminUsersList({query: {q, page}, signal})).data})
}
export function adminUserSearchQuery(q: string) {
	return queryOptions({queryKey: adminKeys.userSearch(q), queryFn: ({signal}) => fetchAllPages(async page => (await adminUsersList({query: {q, page}, signal})).data)})
}
export function adminProjectsQuery(page: number) {
	return queryOptions({queryKey: adminKeys.projectsPage(page), queryFn: async ({signal}) => (await adminProjectsList({query: {page}, signal})).data})
}
export function adminInvitesQuery(page: number) {
	return queryOptions({queryKey: adminKeys.invitesPage(page), queryFn: async ({signal}) => (await adminInviteLinksList({query: {page}, signal})).data})
}
export function adminTeamsQuery(q: string) {
	return queryOptions({queryKey: adminKeys.teams(q), queryFn: ({signal}) => fetchAllPages(async page => (await adminTeamsList({query: {q, page}, signal})).data)})
}
function patchUser(client: QueryClient, updated: AdminUser | undefined) {
	if (!updated?.id) return
	client.setQueriesData<PaginatedAdminUser>({queryKey: adminKeys.users}, current => current && ({...current, items: current.items?.map(u => u.id === updated.id ? updated : u)}))
	client.setQueriesData<User>({queryKey: accountKeys.current}, current => current?.id === updated.id ? {...current, is_admin: updated.is_admin} : current)
}
function invalidateUsers(client: QueryClient) {
	return Promise.all([
		client.invalidateQueries({queryKey: adminKeys.users}),
		client.invalidateQueries({queryKey: ['admin', 'userSearch']}),
		client.invalidateQueries({queryKey: adminKeys.overview}),
	])
}
export function createAdminUserMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (body: CreateUserBodyWritable) => (await adminUsersCreate({body})).data,
		onSettled: (_body, client) => invalidateUsers(client),
		successMessage: data => i18n.global.t('admin.users.createdSuccess', {username: data.username}),
	})
}
export function updateAdminUserMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, is_admin, status}: {id: number, is_admin?: boolean, status?: number}) => {
			const context = captureClientRequestContext()
			let updated: AdminUser | undefined
			if (is_admin !== undefined) updated = (await adminUsersPatchAdmin({path: {id}, body: {is_admin}})).data
			assertClientRequestContext(context)
			if (status !== undefined) updated = (await adminUsersPatchStatus({path: {id}, body: {status}})).data
			return updated
		},
		onSuccess: (updated, _input, client) => patchUser(client, updated),
		onSettled: (_input, client) => invalidateUsers(client),
		successMessage: data => i18n.global.t('admin.users.updatedSuccess', {username: data?.username}),
	})
}
export function setAdminUserPasswordMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, password}: {id: number, password: string}) => (await adminUsersSetPassword({path: {id}, body: {new_password: password}})).data,
		onSuccess: (updated, _input, client) => patchUser(client, updated),
		onSettled: (_input, client) => invalidateUsers(client),
		successMessage: data => i18n.global.t('admin.users.setPasswordSuccess', {username: data.username}),
	})
}
export function resetAdminUserPasswordMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id}: {id: number, username?: string}) => (await adminUsersPasswordResetEmail({path: {id}})).data,
		successMessage: (_data, {username}) => i18n.global.t('admin.users.sendResetEmailSuccess', {username}),
	})
}
export function deleteAdminUserMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, mode}: {id: number, mode: DeleteUserMode, username?: string}) => (await adminUsersDelete({path: {id}, query: {mode}})).data,
		onSettled: (_input, client) => Promise.all([client.invalidateQueries({queryKey: adminKeys.all}), client.invalidateQueries({queryKey: projectKeys.all})]),
		successMessage: (_data, {mode, username}) => i18n.global.t(mode === 'now' ? 'admin.users.deletedSuccess' : 'admin.users.deleteScheduledSuccess', {username}),
	})
}
export function reassignAdminProjectMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, ownerId}: {id: number, ownerId: number}) => (await adminProjectsPatchOwner({path: {id}, body: {owner_id: ownerId}})).data,
		onSettled: (_input, client) => Promise.all([client.invalidateQueries({queryKey: adminKeys.projects}), client.invalidateQueries({queryKey: projectKeys.all})]),
		successMessage: () => i18n.global.t('admin.projects.reassignedSuccess'),
	})
}
export function createAdminInviteMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (body: AdminInviteLinksCreateData['body']) => (await adminInviteLinksCreate({body})).data,
		onSettled: (_input, client) => client.invalidateQueries({queryKey: adminKeys.invites}),
	})
}
export function deleteAdminInviteMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await adminInviteLinksDelete({path: {id}})).data,
		onSettled: (_input, client) => client.invalidateQueries({queryKey: adminKeys.invites}),
		successMessage: () => i18n.global.t('admin.inviteLinks.deleted'),
	})
}
export function useCreateAdminUserMutation() { return useMutation(createAdminUserMutationOptions()) }
export function useUpdateAdminUserMutation() { return useMutation(updateAdminUserMutationOptions()) }
export function useSetAdminUserPasswordMutation() { return useMutation(setAdminUserPasswordMutationOptions()) }
export function useResetAdminUserPasswordMutation() { return useMutation(resetAdminUserPasswordMutationOptions()) }
export function useDeleteAdminUserMutation() { return useMutation(deleteAdminUserMutationOptions()) }
export function useReassignAdminProjectMutation() { return useMutation(reassignAdminProjectMutationOptions()) }
export function useCreateAdminInviteMutation() { return useMutation(createAdminInviteMutationOptions()) }
export function useDeleteAdminInviteMutation() { return useMutation(deleteAdminInviteMutationOptions()) }
