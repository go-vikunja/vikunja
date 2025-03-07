<template>
	<header
		:class="{ 'has-background': background, 'menu-active': menuActive }"
		aria-label="main navigation"
		class="navbar d-print-none"
	>
		<RouterLink
			:to="{ name: 'home' }"
			class="logo-link"
			:aria-label="$t('navigation.overview')"
		>
			<Logo
				width="164"
				height="48"
			/>
		</RouterLink>

		<MenuButton class="menu-button" />

		<div
			v-if="currentProject?.id"
			class="project-title-wrapper"
		>
			<h1 class="project-title">
				{{ currentProject.title === '' ? $t('misc.loading') : getProjectTitle(currentProject) }}
			</h1>

			<BaseButton
				v-if="!isEditorContentEmpty(currentProject.description)"
				:to="{ name: 'project.info', params: { projectId: currentProject.id } }"
				class="project-title-button"
			>
				<span class="tw-sr-only">{{ $t('project.description') }}</span>
				<Icon icon="circle-info" />
			</BaseButton>

			<ProjectSettingsDropdown
				v-if="canWriteCurrentProject && currentProject.id !== -1"
				class="project-title-dropdown"
				:project="currentProject"
			>
				<template #trigger="{ toggleOpen }">
					<BaseButton
						class="project-title-button"
						@click="toggleOpen"
					>
						<span class="tw-sr-only">{{ $t('project.openSettingsMenu') }}</span>
						<Icon
							icon="ellipsis-h"
							class="icon"
						/>
					</BaseButton>
				</template>
			</ProjectSettingsDropdown>
		</div>

		<div class="navbar-end">
			<OpenQuickActions />
			<Notifications />
			<Dropdown>
				<template #trigger="{ toggleOpen, open }">
					<BaseButton
						class="username-dropdown-trigger"
						variant="secondary"
						:shadow="false"
						@click="toggleOpen"
					>
						<img
							:src="authStore.avatarUrl"
							alt=""
							class="avatar"
							width="30"
							height="30"
						>
						<span class="username">{{ authStore.userDisplayName }}</span>
						<span
							class="ml-1 dropdown-icon icon is-small"
							:style="{
								transform: open ? 'rotate(180deg)' : 'rotate(0)',
							}"
						>
							<Icon icon="chevron-down" />
						</span>
					</BaseButton>
				</template>

				<DropdownItem :to="{ name: 'user.settings' }">
					{{ $t('user.settings.title') }}
				</DropdownItem>
				<DropdownItem
					v-if="imprintUrl"
					:href="imprintUrl"
				>
					{{ $t('navigation.imprint') }}
				</DropdownItem>
				<DropdownItem
					v-if="privacyPolicyUrl"
					:href="privacyPolicyUrl"
				>
					{{ $t('navigation.privacy') }}
				</DropdownItem>
				<DropdownItem @click="baseStore.setKeyboardShortcutsActive(true)">
					{{ $t('keyboardShortcuts.title') }}
				</DropdownItem>
				<DropdownItem :to="{ name: 'about' }">
					{{ $t('about.title') }}
				</DropdownItem>
				<DropdownItem @click="authStore.logout()">
					{{ $t('user.auth.logout') }}
				</DropdownItem>
			</Dropdown>
		</div>
	</header>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import { RIGHTS as Rights } from '@/constants/rights'

import ProjectSettingsDropdown from '@/components/project/ProjectSettingsDropdown.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import Notifications from '@/components/notifications/Notifications.vue'
import Logo from '@/components/home/Logo.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import MenuButton from '@/components/home/MenuButton.vue'
import OpenQuickActions from '@/components/misc/OpenQuickActions.vue'

import { getProjectTitle } from '@/helpers/getProjectTitle'
import { isEditorContentEmpty } from '@/helpers/editorContentEmpty'

import { useBaseStore } from '@/stores/base'
import { useConfigStore } from '@/stores/config'
import { useAuthStore } from '@/stores/auth'

const baseStore = useBaseStore()
const currentProject = computed(() => baseStore.currentProject)
const background = computed(() => baseStore.background)
const canWriteCurrentProject = computed(() => baseStore.currentProject?.maxRight > Rights.READ)
const menuActive = computed(() => baseStore.menuActive)

const authStore = useAuthStore()

const configStore = useConfigStore()
const imprintUrl = computed(() => configStore.legal.imprintUrl)
const privacyPolicyUrl = computed(() => configStore.legal.privacyPolicyUrl)
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

	@media screen and (min-width: $tablet) {
		padding-left: 2rem;
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

.project-title-wrapper {
	margin-inline: auto;
	display: flex;
	align-items: center;

	// this makes the truncated text of the project title work
	// inside the flexbox parent
	min-width: 0;

	@media screen and (min-width: $tablet) {
		padding-inline: var(--navbar-gap-width);
	}
}

.project-title {
	font-size: 1rem;
	// We need the following for overflowing ellipsis to work
	text-overflow: ellipsis;
	overflow: hidden;
	white-space: nowrap;

	@media screen and (min-width: $tablet) {
		font-size: 1.75rem;
	}
}

.project-title-dropdown {
	align-self: stretch;

	.project-title-button {
		flex-grow: 1;
	}
}

.project-title-button {
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

	>* {
		min-width: var(--navbar-button-min-width);
	}
}

.username-dropdown-trigger {
	padding-left: .75rem;
	display: inline-flex;
	align-items: center;
	font-size: .85rem;
	font-weight: 700;

	@media screen and (max-width: $tablet) {
		padding-right: .5rem;
	}

	@media screen and (min-width: $tablet) {
		padding-right: .75rem;
	}
}

.username {
	font-family: $vikunja-font;

	@media screen and (max-width: $tablet) {
		display: none;
	}
}

.dropdown-icon {
	transition: transform $transition;
}

.avatar {
	border-radius: 100%;
	vertical-align: middle;
	height: 30px;
	margin-right: .5rem;
}
</style>