import {queryOptions, useMutation} from '@tanstack/vue-query'
import {userExportStatus, userExportRequest, userExportDownload, type UserExportStatus} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {downloadBlob} from '@/helpers/downloadBlob'
import {i18n} from '@/i18n'

export const exportKeys = {status: ['data-export'] as const}

export function dataExportQuery() {
	return queryOptions({
		queryKey: exportKeys.status,
		queryFn: async ({signal}): Promise<UserExportStatus | null> => (await userExportStatus({signal})).data ?? null,
		// Export completes server-side; a remount must not serve a pre-completion status.
		staleTime: 0,
	})
}

export function requestExportMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async (password: string) => (await userExportRequest({body: {password}})).data,
			onSettled: (_input, client) => client.invalidateQueries({queryKey: exportKeys.status}),
			successMessage: () => i18n.global.t('user.export.success'),
		}),
		// Input holds the plaintext password.
		gcTime: 0,
	}
}

export function useRequestExportMutation() {
	return useMutation(requestExportMutationOptions())
}

export function downloadExportMutationOptions() {
	return {
		...contextMutationOptions({
			mutationFn: async (password: string) => {
				const {data} = await userExportDownload({body: {password}, parseAs: 'blob'})
				if (!(data instanceof Blob)) throw new Error('Export response was not a file')
				return data
			},
			onSuccess: blob => downloadBlob(URL.createObjectURL(blob), 'vikunja-export.zip'),
		}),
		// Input holds the plaintext password.
		gcTime: 0,
	}
}

export function useDownloadExportMutation() {
	return useMutation(downloadExportMutationOptions())
}
