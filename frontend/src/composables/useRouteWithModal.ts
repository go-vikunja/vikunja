import {computed, defineAsyncComponent, h, shallowRef, type VNode, watchEffect} from 'vue'
import {useRoute, useRouter, type RouteLocationNormalizedGeneric} from 'vue-router'
import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'

export function useRouteWithModal() {
	const router = useRouter()
	const route = useRoute()
	const backdropView = computed(() => route.fullPath ? window.history.state?.backdropView : undefined)
	const baseStore = useBaseStore()
	const projectStore = useProjectStore()

	const routeWithModal = computed(() => {
		return backdropView.value
			? router.resolve(backdropView.value) as RouteLocationNormalizedGeneric
			: route
	})

	const currentModal = shallowRef<VNode>()
	watchEffect(() => {
		if (!backdropView.value) {
			currentModal.value = undefined
			return
		}

		// this is adapted from vue-router
		// https://github.com/vuejs/vue-router-next/blob/798cab0d1e21f9b4d45a2bd12b840d2c7415f38a/src/RouterView.ts#L125
		const routePropsOption = route.matched[0]?.props.default
		let routeProps = undefined
		if (routePropsOption) {
			if (routePropsOption === true) {
				routeProps = route.params
			} else {
				if (typeof routePropsOption === 'function') {
					routeProps = routePropsOption(route)
				} else {
					routeProps = routePropsOption
				}
			}
		}

		if (typeof routeProps === 'undefined') {
			currentModal.value = undefined
			return
		}

		routeProps.backdropView = backdropView.value

		let component = route.matched[0]?.components?.default

		if (typeof component === 'function') {
			component = defineAsyncComponent(component)
		}

		if (!component) {
			currentModal.value = undefined
			return
		}
		currentModal.value = h(component, routeProps)
	})

	const historyState = computed(() => route.fullPath ? window.history.state : undefined)

	function closeModal() {
		// If the current project was changed because the user moved the currently opened task while coming from kanban,
		// we need to reflect that change in the route when they close the task modal.
		// The last route is only available as resolved string, therefore we need to use a regex for matching here
		const routeMatch = new RegExp('\\/projects\\/\\d+\\/(\\d+)', 'g')
		const match = historyState.value?.back
			? routeMatch.exec(historyState.value.back)
			: null
		if (match !== null && baseStore.currentProject && baseStore.currentProject.id !== 0) {
			let viewId: string | number = match[1]

			if (!viewId) {
				const project = projectStore.projects[baseStore.currentProject.id]
				viewId = project?.views?.[0]?.id
			}

			// Only navigate if we have a valid project and view
			if (baseStore.currentProject.id && viewId) {
				const newRoute = {
					name: 'project.view',
					params: {
						projectId: baseStore.currentProject.id,
						viewId,
					},
				}

				router.push(newRoute)
				return
			}
		}

		// Try browser history first
		if (historyState.value?.back) {
			router.back()
			return
		}

		// Try backdrop view
		const backdropRoute = historyState.value?.backdropView && router.resolve(historyState.value.backdropView)
		if (backdropRoute && backdropRoute.params?.projectId !== '0') {
			router.push(backdropRoute)
			return
		}

		// Fallback to current project or home
		if (baseStore.currentProject && baseStore.currentProject.id !== 0) {
			router.push({
				name: 'project.index',
				params: { projectId: baseStore.currentProject.id },
			})
		} else {
			router.push({ name: 'home' })
		}
	}

	return {routeWithModal, currentModal, closeModal}
}
