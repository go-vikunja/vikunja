<template>
	<header
		:class="{'has-background': background, 'menu-active': menuActive}"
		aria-label="main navigation"
		class="navbar main-theme is-fixed-top d-print-none"
	>
		<router-link :to="{name: 'home'}" class="logo-link">
			<Logo width="164" height="48"/>
		</router-link>
		<MenuButton class="menu-button"/>
		<div class="list-title" ref="listTitle" v-show="currentList.id">
			<template v-if="currentList.id">
				<h1
					:style="{ 'opacity': currentList.title === '' ? '0': '1' }"
					class="title">
					{{ currentList.title === '' ? $t('misc.loading') : getListTitle(currentList) }}
				</h1>

				<BaseButton :to="{name: 'list.info', params: {listId: currentList.id}}" class="info-button">
					<icon icon="circle-info"/>
				</BaseButton>

				<list-settings-dropdown v-if="canWriteCurrentList && currentList.id !== -1" :list="currentList"/>
			</template>
		</div>

		<div class="navbar-end">
			<update/>
			<BaseButton
				@click="openQuickActions"
				class="trigger-button pr-0"
				v-shortcut="'Control+k'"
				:title="$t('keyboardShortcuts.quickSearch')"
			>
				<icon icon="search"/>
			</BaseButton>
			<notifications/>
			<div class="user">
				<dropdown class="is-right" ref="usernameDropdown">
					<template #trigger="{toggleOpen}">
						<x-button
							class="username-dropdown-trigger"
							@click="toggleOpen()"
							variant="secondary"
							:shadow="false"
						>
							<img :src="userAvatar" alt="" class="avatar" width="40" height="40"/>
							<span class="username">{{ userInfo.name !== '' ? userInfo.name : userInfo.username }}</span>
							<span class="icon is-small">
								<icon icon="chevron-down"/>
							</span>
						</x-button>
					</template>

					<dropdown-item
						:to="{name: 'user.settings'}"
					>
						{{ $t('user.settings.title') }}
					</dropdown-item>
					<dropdown-item
						v-if="imprintUrl"
						:href="imprintUrl"
					>
						{{ $t('navigation.imprint') }}
					</dropdown-item>
					<dropdown-item
						v-if="privacyPolicyUrl"
						:href="privacyPolicyUrl"
					>
						{{ $t('navigation.privacy') }}
					</dropdown-item>
					<dropdown-item
						@click="$store.commit('keyboardShortcutsActive', true)"
					>
						{{ $t('keyboardShortcuts.title') }}
					</dropdown-item>
					<dropdown-item
						:to="{name: 'about'}"
					>
						{{ $t('about.title') }}
					</dropdown-item>
					<dropdown-item
						@click="logout()"
					>
						{{ $t('user.auth.logout') }}
					</dropdown-item>
				</dropdown>
			</div>
		</div>
	</header>
</template>

<script setup lang="ts">
import {ref, computed, onMounted, nextTick} from 'vue'
import {useStore} from '@/store'

import {QUICK_ACTIONS_ACTIVE} from '@/store/mutation-types'
import {RIGHTS as Rights} from '@/constants/rights'

import Update from '@/components/home/update.vue'
import ListSettingsDropdown from '@/components/list/list-settings-dropdown.vue'
import Dropdown from '@/components/misc/dropdown.vue'
import DropdownItem from '@/components/misc/dropdown-item.vue'
import Notifications from '@/components/notifications/notifications.vue'
import Logo from '@/components/home/Logo.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import MenuButton from '@/components/home/MenuButton.vue'

import {getListTitle} from '@/helpers/getListTitle'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'

const store = useStore()
const authStore = useAuthStore()
const configStore = useConfigStore()

const userInfo = computed(() => authStore.info)
const userAvatar = computed(() => authStore.avatarUrl)
const currentList = computed(() => store.state.currentList)
const background = computed(() => store.state.background)
const imprintUrl = computed(() => configStore.legal.imprintUrl)
const privacyPolicyUrl = computed(() => configStore.legal.privacyPolicyUrl)
const canWriteCurrentList = computed(() => store.state.currentList.maxRight > Rights.READ)
const menuActive = computed(() => store.state.menuActive)

const usernameDropdown = ref()
const listTitle = ref()
onMounted(async () => {
	await nextTick()
	if (typeof usernameDropdown.value === 'undefined' || typeof listTitle.value === 'undefined') {
		return
	}

	const usernameWidth = usernameDropdown.value.$el.clientWidth
	listTitle.value.style.setProperty('--nav-username-width', `${usernameWidth}px`)
})

function logout() {
	authStore.logout()
}

function openQuickActions() {
	store.commit(QUICK_ACTIONS_ACTIVE, true)
}
</script>

<style lang="scss" scoped>
$vikunja-nav-logo-full-width: 164px;
$user-dropdown-width-mobile: 5rem;

$hamburger-menu-icon-spacing: 1rem;
$hamburger-menu-icon-width: 28px;

.logo-link {
	display: none;
	padding: 0.5rem 0.75rem;

	@media screen and (min-width: $tablet) {
		align-self: stretch;
		display: flex;
		align-items: center;
		padding-left: 2rem;
		margin-right: 1.5rem;
	}
}

.menu-button {
	align-self: stretch;
	margin-right: auto;

	@media screen and (max-width: $tablet) {
		margin-left: $hamburger-menu-icon-spacing;
	}
}

.navbar.main-theme {
	background: var(--site-background);
	justify-content: space-between;
	align-items: center;

	@media screen and (max-width: $desktop) {
		display: flex;
		justify-content: space-between;
	}

	.title {
		margin: 0;
		font-size: 1.75rem;
		text-overflow: ellipsis;
		overflow: hidden;
		white-space: nowrap;
	}

	.navbar-end {
		margin-left: 0;
		align-items: center;
		display: flex;
	}

	@media screen and (max-width: $tablet) {
		&.menu-active {
			z-index: 0;
		}

		.user {
			width: $user-dropdown-width-mobile;

			.username-dropdown-trigger {
				line-height: 1;
				padding: 0 0.25rem;
				height: 1rem;

				.icon {
					width: .5rem;
				}
			}

			.username {
				display: none;
			}
		}
	}
}

.navbar {
	// FIXME: notifications should provide a slot for the icon instead, so that we can style it as we want
	:deep() {
		.trigger-button {
			cursor: pointer;
			color: var(--grey-400);
			padding: .5rem;
			font-size: 1.25rem;
			position: relative;
		}

		> * > .trigger-button {
			width: $navbar-icon-width;
		}
	}

	.user {
		display: flex;
		align-items: center;

		span {
			font-family: $vikunja-font;
		}

		.avatar {
			border-radius: 100%;
			vertical-align: middle;
			height: 40px;
			margin-right: .5rem;
		}

		.username-dropdown-trigger {
			background: none;

			&:focus:not(:active), &:active {
				outline: none !important;
				box-shadow: none !important;
			}
		}
	}
}

.list-title {
	display: flex;
	align-items: center;
	justify-content: center;

	$edit-icon-width: 1rem;

	@media screen and (min-width: $tablet) {
		// We need a fixed width for overflowing ellipsis to work
		--nav-username-width: 0;
		width: calc(100vw - #{$user-dropdown-width-mobile} - #{2 * $hamburger-menu-icon-spacing} - #{$hamburger-menu-icon-width} - #{$edit-icon-width} - #{2 * $navbar-icon-width} - #{$vikunja-nav-logo-full-width} - var(--nav-username-width));
	}

	@media screen and (max-width: $tablet) {
		// We need a fixed width for overflowing ellipsis to work
		width: calc(100vw - #{$user-dropdown-width-mobile} - #{2 * $hamburger-menu-icon-spacing} - #{$hamburger-menu-icon-width} - #{$edit-icon-width} - #{2 * $navbar-icon-width});
	}

	h1 {
		margin: 0;
	}

	:deep(.dropdown-trigger) {
		color: var(--grey-400);
		margin-left: .5rem;
		height: 1rem;
		width: 1rem;
		cursor: pointer;
	}
}

.info-button {
	text-align: center;
	height: 1.25rem;
	line-height: 1.25rem;
	width: 2rem;
	margin-top: .25rem;
	padding: 0 .5rem;
	color: var(--grey-400);
	margin-left: .5rem;
}
</style>