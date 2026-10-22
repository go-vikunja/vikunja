import {queryOptions} from '@tanstack/vue-query'
import {userExportStatus, userExportRequest, userExportDownload, type UserExportStatus} from '@/client/generated'
import {expectBlob} from './blobResponse'
import {contextMutationOptions} from './contextMutation'
import {useSecretMutation} from './secretMutation'
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
	return contextMutationOptions({
		mutationFn: async (password: string) => (await userExportRequest({body: {password}})).data,
		onSettled: (_input, client) => client.invalidateQueries({queryKey: exportKeys.status}),
		successMessage: () => i18n.global.t('user.export.success'),
	})
}

export function useRequestExportMutation() {
	return useSecretMutation(requestExportMutationOptions())
}

export function downloadExportMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (password: string) => {
			const {data} = await userExportDownload({body: {password}, parseAs: 'blob'})
			return expectBlob(data, 'Export')
		},
		onSuccess: blob => downloadBlob(URL.createObjectURL(blob), 'vikunja-export.zip'),
	})
}

export function useDownloadExportMutation() {
	return useSecretMutation(downloadExportMutationOptions())
}
