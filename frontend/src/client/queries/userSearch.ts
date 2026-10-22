import {queryOptions} from '@tanstack/vue-query'
import {projectsUsersSearch, usersSearch} from '@/client/generated'
import {queryClient} from '@/client/queryClient'

export const userSearchKeys = {
	all: ['user-search'] as const,
	global: (query: string) => [...userSearchKeys.all, 'global', query] as const,
	project: (projectId: number, query: string) => [...userSearchKeys.all, 'project', projectId, query] as const,
}

export function userSearchQuery(query: string) {
	return queryOptions({
		queryKey: userSearchKeys.global(query),
		queryFn: async ({signal}) => {
			const {data} = await usersSearch({query: {q: query}, signal})
			return data.items ?? []
		},
	})
}

export function projectUserSearchQuery(projectId: number, query: string) {
	return queryOptions({
		queryKey: userSearchKeys.project(projectId, query),
		queryFn: async ({signal}) => {
			const {data} = await projectsUsersSearch({path: {project: projectId}, query: {q: query}, signal})
			return data.items ?? []
		},
	})
}

export function searchUsers(query: string) {
	return queryClient.fetchQuery(userSearchQuery(query))
}

export function searchProjectUsers(projectId: number, query: string) {
	return queryClient.fetchQuery(projectUserSearchQuery(projectId, query))
}
