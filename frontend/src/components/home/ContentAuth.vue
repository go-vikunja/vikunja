<template>
	<div class="content-auth">
		<BaseButton
			v-show="menuActive"
			class="menu-hide-button d-print-none"
			@click="baseStore.setMenuActive(false)"
		>
			<Icon icon="times" />
		</BaseButton>
		<div
			class="app-container"
			:class="{'has-background': background || blurHash}"
			:style="{'background-image': blurHash && `url(${blurHash})`}"
		>
			<div
				:class="{'is-visible': background}"
				class="app-container-background background-fade-in d-print-none"
				:style="{'background-image': background && `url(${background})`}"
			/>
			<Navigation class="d-print-none" />
			<main
				class="app-content"
				:class="[
					{ 'is-menu-enabled': menuActive },
					$route.name,
				]"
			>
				<BaseButton
					v-show="menuActive"
					class="mobile-overlay d-print-none"
					@click="baseStore.setMenuActive(false)"
				/>

				<QuickActions />

				<RouterView
					v-slot="{ Component }"
					:route="routeWithModal"
				>
					<keep-alive :include="['project.view']">
						<component :is="Component" />
					</keep-alive>
				</RouterView>

				<Modal
					:enabled="typeof currentModal !== 'undefined'"
					variant="scrolling"
					class="task-detail-view-modal"
					@close="closeModal()"
				>
					<component
						:is="currentModal"
						@close="closeModal()"
					/>
				</Modal>

				<BaseButton
					v-shortcut="'Shift+?'"
					class="keyboard-shortcuts-button d-print-none"
					@click="showKeyboardShortcuts()"
				>
					<span class="is-sr-only">{{ $t('keyboardShortcuts.title') }}</span>
					<Icon icon="keyboard" />
				</BaseButton>
			</main>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {watch, computed} from 'vue'
import {useRoute} from 'vue-router'

import Navigation from '@/components/home/Navigation.vue'
import QuickActions from '@/components/quick-actions/QuickActions.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import {useBaseStore} from '@/stores/base'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'

import {useRouteWithModal} from '@/composables/useRouteWithModal'
import {useRenewTokenOnFocus} from '@/composables/useRenewTokenOnFocus'

const {routeWithModal, currentModal, closeModal} = useRouteWithModal()

const baseStore = useBaseStore()
const background = computed(() => baseStore.background)
const blurHash = computed(() => baseStore.blurHash)
const menuActive = computed(() => baseStore.menuActive)

function showKeyboardShortcuts() {
	baseStore.setKeyboardShortcutsActive(true)
}

const route = useRoute()

// FIXME: this is really error prone
// Reset the current project highlight in menu if the current route is not project related.
watch(() => route.name as string, (routeName) => {
	if (
		routeName &&
		(
			[
				'home',
				'teams.index',
				'teams.edit',
				'tasks.range',
				'labels.index',
				'migrate.start',
				'migrate.wunderlist',
				'projects.index',
			].includes(routeName) ||
			routeName.startsWith('user.settings')
		)
	) {
		baseStore.handleSetCurrentProject({project: null})
	}
})

// TODO: Reset the title if the page component does not set one itself

useRenewTokenOnFocus()

const labelStore = useLabelStore()
labelStore.loadAllLabels()

const projectStore = useProjectStore()
projectStore.loadAllProjects()
</script>

<style lang="scss" scoped>
.menu-hide-button {
	position: fixed;
	inset-block-start: 0.5rem;
	inset-inline-end: 0.5rem;
	z-index: 31;
	inline-size: 3rem;
	block-size: 3rem;
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
		color: var(--grey-600);
	}
}

.app-container {
	min-block-size: calc(100vh - 65px);

	@media screen and (max-width: $tablet) {
		padding-block-start: $navbar-height;
	}
}

.app-content {
	display: flow-root;
	z-index: 10;
	position: relative;
	padding: 1.5rem 0.5rem 0;
	// TODO refactor: DRY `transition-timing-function` with `./Navigation.vue`.
	transition: margin-inline-start $transition-duration;

	@media screen and (max-width: $tablet) {
		margin-inline-start: 0;
		margin-inline-end: 0;
		min-block-size: calc(100vh - 4rem);
	}

	@media screen and (min-width: $tablet) {
		padding: $navbar-height + 1.5rem 1.5rem 0 1.5rem;
	}

	&.is-menu-enabled {
		@media screen and (min-width: $tablet) {
			margin-inline-start: $navbar-width;
		}
	}

	// Used to make sure the spinner is always in the middle while loading
	> .loader-container {
		min-block-size: calc(100vh - #{$navbar-height + 1.5rem + 1rem});
	}

	// FIXME: This should be somehow defined inside Card.vue
	.card {
		background: var(--white);
	}
}

.mobile-overlay {
	display: none;
	position: fixed;
	inset-block-start: 0;
	inset-block-end: 0;
	inset-inline-start: 0;
	inset-inline-end: 0;
	block-size: 100vh;
	inline-size: 100vw;
	background: hsla(var(--grey-100-hsl), 0.8);
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
	inset-block-end: calc(1rem - 4px);
	inset-inline-end: 1rem;
	z-index: 4500; // The modal has a z-index of 4000
	color: var(--grey-500);
	transition: color $transition;

	@media screen and (max-width: $tablet) {
		display: none;
	}
}

.content-auth {
	position: relative;
	z-index: 1;
}
</style>
