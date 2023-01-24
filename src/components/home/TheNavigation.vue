<template>
	<header
		:class="{'has-background': background, 'menu-active': menuActive}"
		aria-label="main navigation"
		class="navbar d-print-none"
	>
		<router-link :to="{name: 'home'}" class="logo-link">
			<Logo width="164" height="48"/>
		</router-link>

		<MenuButton class="menu-button"/>

		<div
			v-if="currentList.id"
			class="list-title-wrapper"
		>
			<h1 class="list-title">{{ currentList.title === '' ? $t('misc.loading') : getListTitle(currentList) }}</h1>
			
			<BaseButton :to="{name: 'list.info', params: {listId: currentList.id}}" class="list-title-button">
				<icon icon="circle-info"/>
			</BaseButton>

			<list-settings-dropdown
				v-if="canWriteCurrentList && currentList.id !== -1"
				class="list-title-dropdown"
				:list="currentList"
			>
				<template #trigger="{toggleOpen}">
					<BaseButton class="list-title-button" @click="toggleOpen">
						<icon icon="ellipsis-h" class="icon"/>
					</BaseButton>
				</template>
			</list-settings-dropdown>
		</div>

		<div class="navbar-end">
			<BaseButton
				@click="openQuickActions"
				class="trigger-button"
				v-shortcut="'Control+k'"
				:title="$t('keyboardShortcuts.quickSearch')"
			>
				<icon icon="search"/>
			</BaseButton>
			<Notifications />
			<dropdown>
				<template #trigger="{toggleOpen, open}">
					<BaseButton
						class="username-dropdown-trigger"
						@click="toggleOpen"
						variant="secondary"
						:shadow="false"
					>
						<img :src="authStore.avatarUrl" alt="" class="avatar" width="40" height="40"/>
						<span class="username">{{ authStore.userDisplayName }}</span>
						<span class="icon is-small" :style="{
							transform: open ? 'rotate(180deg)' : 'rotate(0)',
						}">
							<icon icon="chevron-down"/>
						</span>
					</BaseButton>
				</template>

				<dropdown-item :to="{name: 'user.settings'}">
					{{ $t('user.settings.title') }}
				</dropdown-item>
				<dropdown-item v-if="imprintUrl" :href="imprintUrl">
					{{ $t('navigation.imprint') }}
				</dropdown-item>
				<dropdown-item v-if="privacyPolicyUrl" :href="privacyPolicyUrl">
					{{ $t('navigation.privacy') }}
				</dropdown-item>
				<dropdown-item @click="baseStore.setKeyboardShortcutsActive(true)">
					{{ $t('keyboardShortcuts.title') }}
				</dropdown-item>
				<dropdown-item :to="{name: 'about'}">
					{{ $t('about.title') }}
				</dropdown-item>
				<dropdown-item @click="authStore.logout()">
					{{ $t('user.auth.logout') }}
				</dropdown-item>
			</dropdown>
		</div>
	</header>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import {RIGHTS as Rights} from '@/constants/rights'

import ListSettingsDropdown from '@/components/list/list-settings-dropdown.vue'
import Dropdown from '@/components/misc/dropdown.vue'
import DropdownItem from '@/components/misc/dropdown-item.vue'
import Notifications from '@/components/notifications/notifications.vue'
import Logo from '@/components/home/Logo.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import MenuButton from '@/components/home/MenuButton.vue'

import {getListTitle} from '@/helpers/getListTitle'

import {useBaseStore} from '@/stores/base'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'

const baseStore = useBaseStore()
const currentList = computed(() => baseStore.currentList)
const background = computed(() => baseStore.background)
const canWriteCurrentList = computed(() => baseStore.currentList.maxRight > Rights.READ)
const menuActive = computed(() => baseStore.menuActive)

const authStore = useAuthStore()

const configStore = useConfigStore()
const imprintUrl = computed(() => configStore.legal.imprintUrl)
const privacyPolicyUrl = computed(() => configStore.legal.privacyPolicyUrl)

function openQuickActions() {
	baseStore.setQuickActionsActive(true)
}
</script>

<style lang="scss" scoped>
$user-dropdown-width-mobile: 5rem;

.navbar {
	--navbar-button-min-width: 40px;
	--navbar-gap-width: 1rem;
	--navbar-icon-size: 1.25rem;

	position: fixed;
	top: 0;
	left: 0;
	right: 0;

	display: flex;
	justify-content: space-between;
	gap: var(--navbar-gap-width);

	background: var(--site-background);

	@media screen and (max-width: $tablet) {
		padding-right: .5rem;
	}

	@media screen and (min-width: $tablet) {
		padding-left: 2rem;
		padding-right: 1rem;
		align-items: stretch;
	}

	&.menu-active {
		@media screen and (max-width: $tablet) {
			z-index: 0;
		}
	}

	// FIXME: notifications should provide a slot for the icon instead, so that we can style it as we want
	:deep() {
		.trigger-button {
			color: var(--grey-400);
			font-size: var(--navbar-icon-size);
		}
	}
}

.logo-link {
	display: none;
	
	@media screen and (min-width: $tablet) {
		align-self: stretch;
		display: flex;
		align-items: center;
		margin-right: .5rem;
	}
}

.menu-button {
	margin-right: auto;
	align-self: stretch;
	flex: 0 0 auto;

	@media screen and (max-width: $tablet) {
		margin-left: 1rem;
	}
}

.list-title-wrapper {
	margin-inline: auto;
	display: flex;
	align-items: center;

	// this makes the truncated text of the list title work
	// inside the flexbox parent
	min-width: 0;

	@media screen and (min-width: $tablet) {
		padding-inline: var(--navbar-gap-width);
	}
}

.list-title {
	font-size: 1rem;
	// We need the following for overflowing ellipsis to work
	text-overflow: ellipsis;
	overflow: hidden;
	white-space: nowrap;

	@media screen and (min-width: $tablet) {
		font-size: 1.75rem;
	}
}

.list-title-dropdown {
	align-self: stretch;

	.list-title-button {
		flex-grow: 1;
	}
}

.list-title-button {
	align-self: stretch;
	min-width: var(--navbar-button-min-width);
	display: flex;
	place-items: center;
	justify-content: center;
	font-size: var(--navbar-icon-size);
	color: var(--grey-400);
}

.navbar-end {
	margin-left: auto;
	flex: 0 0 auto;
	display: flex;
	align-items: stretch;

	> * {
		min-width: var(--navbar-button-min-width);
	}
}

.username-dropdown-trigger {
	padding-left: 1rem;
	display: inline-flex;
	align-items: center;
	text-transform: uppercase;
	font-size: .85rem;
	font-weight: 700;
}

.username {
	font-family: $vikunja-font;

	@media screen and (max-width: $tablet) {
		display: none;
	}
}

.avatar {
	border-radius: 100%;
	vertical-align: middle;
	height: 40px;
	margin-right: .5rem;
}
</style>