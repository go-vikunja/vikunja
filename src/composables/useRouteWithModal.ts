import { computed, shallowRef, watchEffect, h, type VNode } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export function useRouteWithModal() {
	const router = useRouter()
	const route = useRoute()
	const backdropView = computed(() => route.fullPath && window.history.state.backdropView)

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

		// logic from vue-router
		// https://github.com/vuejs/vue-router-next/blob/798cab0d1e21f9b4d45a2bd12b840d2c7415f38a/src/RouterView.ts#L125
		const routePropsOption = route.matched[0]?.props.default
		const routeProps = routePropsOption
			? routePropsOption === true
				? route.params
				: typeof routePropsOption === 'function'
					? routePropsOption(route)
					: routePropsOption
			: null

		const component = route.matched[0]?.components?.default

		if (!component) {
			currentModal.value = undefined
			return
		}
		currentModal.value = h(component, routeProps)
	})

	function closeModal() {
		const historyState = computed(() => route.fullPath && window.history.state)

		if (historyState.value) {
			router.back()
		} else {
			const backdropRoute = historyState.value?.backdropView && router.resolve(historyState.value.backdropView)
			router.push(backdropRoute)
		}
	}

	return {routeWithModal, currentModal, closeModal}
}