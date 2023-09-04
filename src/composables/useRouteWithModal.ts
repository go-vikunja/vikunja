import {computed, shallowRef, watchEffect, h, type VNode} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useBaseStore} from '@/stores/base'

export function useRouteWithModal() {
	const router = useRouter()
	const route = useRoute()
	const backdropView = computed(() => route.fullPath && window.history.state.backdropView)
	const baseStore = useBaseStore()

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
		const routeProps = routePropsOption
			? routePropsOption === true
				? route.params
				: typeof routePropsOption === 'function'
					? routePropsOption(route)
					: routePropsOption
			: {}

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

	function closeModal() {
		const historyState = computed(() => route.fullPath && window.history.state)

		// If the current project was changed because the user moved the currently opened task while coming from kanban,
		// we need to reflect that change in the route when they close the task modal.
		// The last route is only available as resolved string, therefore we need to use a regex for matching here
		const kanbanRouteMatch = new RegExp('\\/projects\\/\\d+\\/kanban', 'g')
		const kanbanRouter = {name: 'project.kanban', params: {projectId: baseStore.currentProject?.id}}
		if (kanbanRouteMatch.test(historyState.value.back)
			&& baseStore.currentProject
			&& historyState.value.back !== router.resolve(kanbanRouter).fullPath) {
			router.push(kanbanRouter)
			return
		}

		if (historyState.value) {
			router.back()
		} else {
			const backdropRoute = historyState.value?.backdropView && router.resolve(historyState.value.backdropView)
			router.push(backdropRoute)
		}
	}

	return {routeWithModal, currentModal, closeModal}
}