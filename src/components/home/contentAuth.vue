<template>
	<div>
		<a @click="$store.commit('menuActive', false)" class="menu-hide-button" v-if="menuActive">
			<icon icon="times" />
		</a>
		<div
			:class="{'has-background': background}"
			:style="{'background-image': background && `url(${background})`}"
			class="app-container"
		>
			<navigation/>
			<main
				:class="[
					{ 'is-menu-enabled': menuActive },
					$route.name,
				]"
				class="app-content"
			>
				<a @click="$store.commit('menuActive', false)" class="mobile-overlay" v-if="menuActive"></a>

				<quick-actions/>


				<router-view :route="routeWithModal" v-slot="{ Component }">
					<keep-alive :include="['list.list', 'list.gantt', 'list.table', 'list.kanban']">
						<component :is="Component" />
					</keep-alive>
				</router-view>

				<transition name="modal">
					<component v-if="currentModal" :is="currentModal" />
				</transition>

				<a
					class="keyboard-shortcuts-button"
					@click="showKeyboardShortcuts()"
					v-shortcut="'?'"
				>
					<icon icon="keyboard"/>
				</a>
			</main>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {watch, computed, shallowRef, watchEffect, VNode, h} from 'vue'
import {useStore} from 'vuex'
import {useRoute, useRouter} from 'vue-router'
import {useEventListener} from '@vueuse/core'

import {CURRENT_LIST, KEYBOARD_SHORTCUTS_ACTIVE, MENU_ACTIVE} from '@/store/mutation-types'
import Navigation from '@/components/home/navigation.vue'
import QuickActions from '@/components/quick-actions/quick-actions.vue'

function useRouteWithModal() {
	const router = useRouter()
	const route = useRoute()
	const historyState = computed(() => route.fullPath && window.history.state)

	const routeWithModal = computed(() => {
		if (historyState.value.backdropView) {
			return router.resolve(historyState.value.backdropView)
		} else {
			return route
		}
	})

	const currentModal = shallowRef<VNode>()
	watchEffect(() => {
		const hasBackdropView = historyState.value.backdropView

		if (!hasBackdropView) {
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

		currentModal.value = h(
			route.matched[0]?.components.default,
			routeProps,
		)
	})

	return { routeWithModal, currentModal }
}

const { routeWithModal, currentModal } = useRouteWithModal()

const store = useStore()

const background = computed(() => store.state.background)
const menuActive = computed(() => store.state.menuActive)

function showKeyboardShortcuts() {
	store.commit(KEYBOARD_SHORTCUTS_ACTIVE, true)
}

const route = useRoute()

// hide menu on mobile
watch(() => route.fullPath, () => window.innerWidth < 769 && store.commit(MENU_ACTIVE, false))

// FIXME: this is really error prone
// Reset the current list highlight in menu if the current route is not list related.
watch(() => route.name as string, (routeName) => {
	if (
		routeName &&
		(
			[
				'home',
				'namespace.edit',
				'teams.index',
				'teams.edit',
				'tasks.range',
				'labels.index',
				'migrate.start',
				'migrate.wunderlist',
				'namespaces.index',
			].includes(routeName) || 
			routeName.startsWith('user.settings')
		)
	) {
		store.dispatch(CURRENT_LIST, null)
	}
})

// TODO: Reset the title if the page component does not set one itself

function useRenewTokenOnFocus() {
	const router = useRouter()
	
	const userInfo = computed(() => store.state.auth.info)
	const authenticated = computed(() => store.state.auth.authenticated)

	// Try renewing the token every time vikunja is loaded initially
	// (When opening the browser the focus event is not fired)
	store.dispatch('auth/renewToken')

	// Check if the token is still valid if the window gets focus again to maybe renew it
	useEventListener('focus', () => {
		if (!authenticated.value) {
			return
		}

		const expiresIn = (userInfo.value !== null ? userInfo.value.exp : 0) - +new Date() / 1000

		// If the token expiry is negative, it is already expired and we have no choice but to redirect
		// the user to the login page
		if (expiresIn < 0) {
			store.dispatch('auth/checkAuth')
			router.push({name: 'user.login'})
			return
		}

		// Check if the token is valid for less than 60 hours and renew if thats the case
		if (expiresIn < 60 * 3600) {
			store.dispatch('auth/renewToken')
			console.debug('renewed token')
		}
	})
}

useRenewTokenOnFocus()
store.dispatch('labels/loadAllLabels')
</script>

<style lang="scss" scoped>
.menu-hide-button {
	position: fixed;
	top: 0.5rem;
	right: 0.5rem;
	z-index: 31;
	width: 3rem;
	height: 3rem;
	display: flex;
	justify-content: center;
	align-items: center;
	font-size: 2rem;
	color: var(--grey-400);
	line-height: 1;
	transition: all $transition;

	@media screen and (min-width: $tablet) {
		display: none;
	}

	&:hover,
	&:focus {
		height: 1rem;
		color: var(--grey-600);
	}
}

.app-container {
  min-height: calc(100vh - 65px);

  @media screen and (max-width: $tablet) {
    padding-top: $navbar-height;
  }

  .app-content {
    padding: $navbar-height + 1.5rem 1.5rem 1rem 1.5rem;
    z-index: 2;

    @media screen and (max-width: $tablet) {
      margin-left: 0;
      padding-top: 1.5rem;
      min-height: calc(100vh - 4rem);
    }

    &.is-menu-enabled {
      margin-left: $navbar-width;

      @media screen and (max-width: $tablet) {
        min-width: 100%;
        margin-left: 0;
      }
    }

	&.task\.detail {
		padding-left: 0;
		padding-right: 0;
	}

    .card {
      background: var(--white);
    }
  }
}

.mobile-overlay {
	display: none;
	position: fixed;
	top: 0;
	bottom: 0;
	left: 0;
	right: 0;
	background: rgba(250, 250, 250, 0.8);
	z-index: 5;
	opacity: 0;
	transition: all $transition;

	@media screen and (max-width: $tablet) {
		display: block;
		opacity: 1;
	}
}

.keyboard-shortcuts-button {
	position: fixed;
	bottom: calc(1rem - 4px);
	right: 1rem;
	z-index: 4500; // The modal has a z-index of 4000

	color: var(--grey-500);
	transition: color $transition;
	
	@media screen and (max-width: $tablet) {
		display: none;
	}
}

@include modal-transition();
</style>