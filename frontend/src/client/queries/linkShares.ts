import {queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'
import {sharesCreate, sharesDelete, sharesList} from '@/client/generated'
import type {LinkSharing, LinkSharingWritable} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {fetchAllPages} from './fetchAllPages'
import {API_MAX_PER_PAGE} from './pagination'
import {normalizeSharePermission} from './projectShares'
import {i18n} from '@/i18n'

export const linkShareKeys = {
	list: (projectId: number) => ['link-shares', projectId] as const,
}

export function createLinkShareDraft(share: LinkSharingWritable = {}): Required<LinkSharingWritable> {
	return {name: share.name ?? '', password: share.password ?? '', permission: normalizeSharePermission(share.permission)}
}

export function linkSharesQuery(projectId: number) {
	return queryOptions({
		queryKey: linkShareKeys.list(projectId),
		queryFn: ({signal}) => fetchAllPages(async page => (await sharesList({path: {project: projectId}, query: {page, per_page: API_MAX_PER_PAGE}, signal})).data),
	})
}

function invalidateLinkShares(client: QueryClient, projectId: number) {
	return client.invalidateQueries({queryKey: linkShareKeys.list(projectId)})
}

export function createLinkShareMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async ({projectId, share}: {projectId: number, share: LinkSharingWritable}) => (await sharesCreate({path: {project: projectId}, body: createLinkShareDraft(share)})).data,
			onSuccess: (created, {projectId}, client) => {
				client.setQueryData<LinkSharing[]>(linkShareKeys.list(projectId), current => current ? [...current, created] : current)
			},
			onSettled: ({projectId}, client) => invalidateLinkShares(client, projectId),
			successMessage: () => i18n.global.t('project.share.links.createSuccess'),
		}),
		// Input holds the plaintext password.
		gcTime: 0,
	}
}

export function deleteLinkShareMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({projectId, id}: {projectId: number, id: number}) => { await sharesDelete({path: {project: projectId, share: id}}) },
		onSuccess: (_data, {projectId, id}, client) => {
			client.setQueryData<LinkSharing[]>(linkShareKeys.list(projectId), current => current?.filter(share => share.id !== id))
		},
		onSettled: ({projectId}, client) => invalidateLinkShares(client, projectId),
		successMessage: () => i18n.global.t('project.share.links.deleteSuccess'),
	})
}

export const useCreateLinkShareMutation = () => useMutation(createLinkShareMutationOptions())
export const useDeleteLinkShareMutation = () => useMutation(deleteLinkShareMutationOptions())
