import {queryOptions} from '@tanstack/vue-query'
import {linkPreviewRead, type Preview} from '@/client/generated'

export const linkPreviewKeys = {
	all: ['linkPreviews'] as const,
	detail: (url: string) => [...linkPreviewKeys.all, url] as const,
}

export type LinkPreviewResponse = Required<Omit<Preview, '$schema'>>

export function normalizeLinkPreview(preview: Preview): LinkPreviewResponse {
	return {
		url: preview.url ?? '',
		title: preview.title ?? '',
		description: preview.description ?? '',
		image: preview.image ?? '',
		site_name: preview.site_name ?? '',
		favicon: preview.favicon ?? '',
	}
}

export function linkPreviewQuery(url: string) {
	return queryOptions({
		queryKey: linkPreviewKeys.detail(url),
		queryFn: async ({signal}) => {
			const {data} = await linkPreviewRead({
				query: {url},
				signal,
			})
			return normalizeLinkPreview(data)
		},
		staleTime: 30 * 60 * 1000,
		gcTime: 30 * 60 * 1000,
		retry: false,
		retryOnMount: false,
		refetchOnWindowFocus: false,
		meta: {handlesError: true},
	})
}
