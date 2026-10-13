import {useRoute, useRouter} from 'vue-router'

// router.back() does nothing when the page is the tab's first history entry (new tab, bookmark).
export function useGoBack() {
	const router = useRouter()
	const route = useRoute()

	return function goBack() {
		if (router.options.history.state.back) {
			router.back()
			return
		}

		const projectId = Number(route.params.projectId)
		router.push(projectId
			? {name: 'project.index', params: {projectId}}
			: {name: 'home'})
	}
}
