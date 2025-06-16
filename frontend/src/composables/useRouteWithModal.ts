import {computed, h, shallowRef, type VNode, watchEffect} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'

export function useRouteWithModal() {
	const router = useRouter()
	const route = useRoute()
	const backdropView = computed(() => route.fullPath ? window.history.state.backdropView : undefined)
	const baseStore = useBaseStore()
	const projectStore = useProjectStore()

	const routeWithModal = computed(() => {
		return backdropView.value
			? router.resolve(backdropView.value)
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

		const component = route.matched[0]?.components?.default

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
		if (match !== null && baseStore.currentProject) {
			let viewId: string | number = match[1]

			if (!viewId) {
				viewId = projectStore.projects[baseStore.currentProject?.id].views[0]?.id
			}

			const newRoute = {
				name: 'project.view',
				params: {
					projectId: baseStore.currentProject?.id,
					viewId,
				},
			}
			router.push(newRoute)
			return
		}

		if (historyState.value) {
			router.back()
		} else {
			const backdropRoute = historyState.value?.backdropView && router.resolve(historyState.value.backdropView)
			if (backdropRoute) {
				router.push(backdropRoute)
			}
		}
	}

	return {routeWithModal, currentModal, closeModal}
}
