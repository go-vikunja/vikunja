import {mutationOptions, queryOptions} from '@tanstack/vue-query'
import type {QueryClient} from '@tanstack/vue-query'

import {client} from '@/client/generated/client.gen'

// These talk to /api/v2/manage/*, the people area. The operations are not in the generated SDK
// yet: run `mage generate:frontend-client` (needs Go) and replace the bodies of the functions
// below with the generated `manage*` operations. Until then they go through the same generated
// runtime client, so authentication, token refresh and the identity fencing in client/http.ts
// apply unchanged. The wire types below mirror the Go structs in pkg/models/manage_users.go
// and pkg/routes/api/v2/manage_*.go.

export type ManageUserSource = 'local' | 'entra' | 'import' | 'ldap' | 'other'

export interface ManagedUser {
	id: number
	username: string
	name: string
	email: string
	language: string
	job_title: string
	department: string
	status: number
	is_admin: boolean
	source: ManageUserSource
	issuer: string
	import_id?: string
	must_change_password: boolean
	profile_editable: boolean
	created: string
}

export interface ManagedUsersPage {
	items: ManagedUser[]
	total: number
	page: number
	per_page: number
	total_pages: number
}

export interface ManageUsersFilter {
	page: number
	q?: string
	status?: number
	source?: ManageUserSource
}

export interface TransferableProject {
	id: number
	title: string
	is_archived: boolean
	parent_project_id: number
}

export interface ImportCounts {
	created: number
	failed: number
	updated: number
	reenabled: number
	disabled: number
	unchanged: number
	exempt: number
	skipped_invalid: number
	skipped_duplicate: number
}

export interface ImportEntry {
	username: string
	name: string
	email: string
}

export interface ImportDetails {
	create: ImportEntry[]
	update: ImportEntry[]
	disable: ImportEntry[]
	truncated: boolean
}

export interface ImportLastRun {
	trigger: 'schedule' | 'manual'
	actor?: string
	finished_at: string
	dry_run: boolean
	counts?: ImportCounts
	error?: string
}

export interface ImportStatus {
	enabled: boolean
	schedule: string
	dry_run: boolean
	tenant_set: boolean
	file_exists: boolean
	file_size: number
	file_modified?: string
	running: boolean
	last_run?: ImportLastRun
}

export interface ImportPreview {
	counts: ImportCounts
	details?: ImportDetails
	blocked?: string
}

export interface ImportUploadResult {
	rows: number
	skipped_invalid: number
	skipped_duplicate: number
}

export const MANAGE_PAGE_SIZE = 50

export const manageKeys = {
	all: ['manage'] as const,
	users: () => [...manageKeys.all, 'users'] as const,
	userList: (filter: ManageUsersFilter) => [...manageKeys.users(), 'list', filter] as const,
	userProjects: (id: number) => [...manageKeys.users(), 'projects', id] as const,
	importStatus: () => [...manageKeys.all, 'user-import', 'status'] as const,
}

async function request<T>(method: 'get' | 'post' | 'put' | 'patch' | 'delete', url: string, options: Record<string, unknown> = {}): Promise<T> {
	const result = await client[method]({url, ...options, throwOnError: true} as never) as {data: T}
	return result.data
}

export function manageUsersQuery(filter: ManageUsersFilter) {
	return queryOptions({
		queryKey: manageKeys.userList(filter),
		queryFn: () => request<ManagedUsersPage>('get', '/manage/users', {
			query: {
				page: filter.page,
				per_page: MANAGE_PAGE_SIZE,
				...(filter.q ? {q: filter.q} : {}),
				...(filter.status !== undefined ? {status: filter.status} : {}),
				...(filter.source ? {source: filter.source} : {}),
			},
		}),
		staleTime: 30 * 1000,
	})
}

export function manageUserProjectsQuery(id: number) {
	return queryOptions({
		queryKey: manageKeys.userProjects(id),
		queryFn: async () => (await request<{items: TransferableProject[]}>('get', `/manage/users/${id}/projects`)).items,
	})
}

export function importStatusQuery(pollWhileRunning = true) {
	return queryOptions({
		queryKey: manageKeys.importStatus(),
		queryFn: () => request<ImportStatus>('get', '/manage/user-import'),
		// The run is a background job: poll until it finishes.
		refetchInterval: query => pollWhileRunning && query.state.data?.running ? 2000 : false,
	})
}

// Every change can alter the list (status, source, counts), so mutations end by invalidating it.
// Cache writes happen only in the callbacks, with the client they are given.
function invalidateUsers(queryClient: QueryClient) {
	return queryClient.invalidateQueries({queryKey: manageKeys.users()})
}

export interface UpdateProfileInput {
	id: number
	name: string
	email: string
	language?: string
}

export function updateManagedUserMutation() {
	return mutationOptions({
		mutationFn: ({id, ...body}: UpdateProfileInput) => request<ManagedUser>('put', `/manage/users/${id}`, {
			body,
			headers: {'Content-Type': 'application/json'},
		}),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateUsers(context.client),
	})
}

export interface CreateManagedUserInput {
	username: string
	email: string
	password: string
	name?: string
	language?: string
	is_admin?: boolean
	skip_email_confirm?: boolean
	require_change?: boolean
}

export function createManagedUserMutation() {
	return mutationOptions({
		mutationFn: (body: CreateManagedUserInput) => request<ManagedUser>('post', '/manage/users', {
			body,
			headers: {'Content-Type': 'application/json'},
		}),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateUsers(context.client),
	})
}

export function setManagedUserAdminMutation() {
	return mutationOptions({
		mutationFn: ({id, is_admin}: {id: number, is_admin: boolean}) => request<ManagedUser>('patch', `/manage/users/${id}/admin`, {
			body: {is_admin},
			headers: {'Content-Type': 'application/json'},
		}),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateUsers(context.client),
	})
}

export function setManagedUserStatusMutation() {
	return mutationOptions({
		mutationFn: ({id, status}: {id: number, status: number}) => request<ManagedUser>('patch', `/manage/users/${id}/status`, {
			body: {status},
			headers: {'Content-Type': 'application/json'},
		}),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateUsers(context.client),
	})
}

export function setManagedUserPasswordMutation() {
	return mutationOptions({
		mutationFn: ({id, new_password, require_change}: {id: number, new_password: string, require_change: boolean}) =>
			request<ManagedUser>('patch', `/manage/users/${id}/password`, {
				body: {new_password, require_change},
				headers: {'Content-Type': 'application/json'},
			}),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateUsers(context.client),
	})
}

export function deleteManagedUserMutation() {
	return mutationOptions({
		mutationFn: ({id, mode}: {id: number, mode: 'now' | 'scheduled'}) => request<void>('delete', `/manage/users/${id}`, {query: {mode}}),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateUsers(context.client),
	})
}

export function transferProjectsMutation() {
	return mutationOptions({
		mutationFn: ({id, new_owner_id, project_ids}: {id: number, new_owner_id: number, project_ids?: number[]}) =>
			request<{transferred: number}>('post', `/manage/users/${id}/transfer-projects`, {
				body: {new_owner_id, ...(project_ids ? {project_ids} : {})},
				headers: {'Content-Type': 'application/json'},
			}),
		onSettled: (_data, _error, vars, _ctx, context) => context.client.invalidateQueries({queryKey: manageKeys.userProjects(vars.id)}),
	})
}

export function uploadUserListMutation() {
	return mutationOptions({
		mutationFn: (file: File) => {
			const body = new FormData()
			body.append('list', file)
			// A null content type lets the browser set the multipart boundary.
			return request<ImportUploadResult>('put', '/manage/user-import/file', {
				body,
				bodySerializer: null,
				headers: {'Content-Type': null},
			})
		},
		onSettled: (_data, _error, _vars, _ctx, context) => context.client.invalidateQueries({queryKey: manageKeys.importStatus()}),
	})
}

export function previewUserImportMutation() {
	return mutationOptions({
		mutationFn: () => request<ImportPreview>('post', '/manage/user-import/preview'),
	})
}

export function runUserImportMutation() {
	return mutationOptions({
		mutationFn: () => request<{message: string}>('post', '/manage/user-import/run'),
		onSettled: (_data, _error, _vars, _ctx, context) => context.client.invalidateQueries({queryKey: manageKeys.importStatus()}),
	})
}
