<template>
	<aside
		:class="{'is-active': baseStore.menuActive}"
		class="menu-container"
	>
		<nav class="menu top-menu">
			<RouterLink
				:to="{name: 'home'}"
				class="logo"
			>
				<Logo
					width="164"
					height="48"
				/>
			</RouterLink>
			<menu class="menu-list other-menu-items">
				<li>
					<RouterLink
						v-shortcut="'g o'"
						:to="{ name: 'home'}"
					>
						<span class="menu-item-icon icon">
							<Icon icon="calendar" />
						</span>
						{{ $t('navigation.overview') }}
					</RouterLink>
				</li>
				<li>
					<RouterLink
						v-shortcut="'g u'"
						:to="{ name: 'tasks.range'}"
					>
						<span class="menu-item-icon icon">
							<Icon :icon="['far', 'calendar-alt']" />
						</span>
						{{ $t('navigation.upcoming') }}
					</RouterLink>
				</li>
				<li>
					<RouterLink
						v-shortcut="'g p'"
						:to="{ name: 'projects.index'}"
					>
						<span class="menu-item-icon icon">
							<Icon icon="layer-group" />
						</span>
						{{ $t('project.projects') }}
					</RouterLink>
				</li>
				<li>
					<RouterLink
						v-shortcut="'g a'"
						:to="{ name: 'labels.index'}"
					>
						<span class="menu-item-icon icon">
							<Icon icon="tags" />
						</span>
						{{ $t('label.title') }}
					</RouterLink>
				</li>
				<li>
					<RouterLink
						v-shortcut="'g m'"
						:to="{ name: 'teams.index'}"
					>
						<span class="menu-item-icon icon">
							<Icon icon="users" />
						</span>
						{{ $t('team.title') }}
					</RouterLink>
				</li>
			</menu>
		</nav>

		<Loading
			v-if="projectStore.isLoading"
			variant="small"
		/>
		<nav
			v-else
			class="menu"
		>
			<ProjectsNavigation
				v-if="favoriteProjects?.length"
				:model-value="favoriteProjects" 
				:can-edit-order="false"
				:can-collapse="false"
			/>

			<ProjectsNavigation
				v-if="savedFilterProjects?.length"
				:model-value="savedFilterProjects"
				:can-edit-order="false"
				:can-collapse="false"
			/>

			<ProjectsNavigation
				v-if="projects?.length"
				:model-value="projects"
				:can-edit-order="true"
				:can-collapse="true"
				:level="1"
			/>
		</nav>

		<PoweredByLink
			class="mt-auto"
			utm-medium="navigation"
		/>
	</aside>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'

import PoweredByLink from '@/components/home/PoweredByLink.vue'
import Logo from '@/components/home/Logo.vue'
import Loading from '@/components/misc/Loading.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const {
	notArchivedRootProjects: projects,
	favoriteProjects,
	savedFilterProjects,
} = storeToRefs(projectStore)
</script>

<style lang="scss" scoped>
.logo {
	padding-left: 1rem;
	margin-right: 1rem;
	margin-bottom: 1rem;

	@media screen and (min-width: $tablet) {
		display: none;
	}
}

.menu-container {
	display: flex;
	flex-direction: column;
	background: var(--site-background);
	color: $vikunja-nav-color;
	padding: 1rem 0;
	transition: transform $transition-duration ease-in;
	position: fixed;
	top: $navbar-height;
	bottom: 0;
	left: 0;
	transform: translateX(-100%);
	width: $navbar-width;
	overflow-y: auto;

	@media screen and (max-width: $tablet) {
		top: 0;
		width: 70vw;
		z-index: 20;
	}

	&.is-active {
		transform: translateX(0);
		transition: transform $transition-duration ease-out;
	}
}

.top-menu .menu-list {
	li {
		font-weight: 600;
		font-family: $vikunja-font;
	}

	.list-menu-link,
	li > a {
		padding-left: 2rem;
		display: inline-block;

		.icon {
			padding-bottom: .25rem;
		}
	}
}

.menu-list + .menu-list {
	padding-top: math.div($navbar-padding, 2);
}
</style>
