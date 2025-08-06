<template>
	<aside
		:class="{'is-active': baseStore.menuActive}"
		class="menu-container"
	>
		<nav class="menu top-menu">
			<RouterLink
				:to="{name: 'home'}"
				class="logo"
				:aria-label="$t('navigation.overview')"
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
		<template v-else>
			<nav
				v-if="favoriteProjects.length"
				class="menu"
			>
				<ProjectsNavigation
					:model-value="favoriteProjects" 
					:can-edit-order="false"
					:can-collapse="false"
				/>
			</nav>
			
			<nav
				v-if="savedFilterProjects.length"
				class="menu"
			>
				<ProjectsNavigation
					:model-value="savedFilterProjects"
					:can-edit-order="false"
					:can-collapse="false"
				/>
			</nav>

			<nav class="menu">
				<ProjectsNavigation
					:model-value="projects"
					:can-edit-order="true"
					:can-collapse="true"
				/>
			</nav>
		</template>

		<PoweredByLink
			class="mbs-auto"
			utm-medium="navigation"
		/>
	</aside>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import PoweredByLink from '@/components/home/PoweredByLink.vue'
import Logo from '@/components/home/Logo.vue'
import Loading from '@/components/misc/Loading.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'

const baseStore = useBaseStore()
const projectStore = useProjectStore()

const projects = computed(() => projectStore.notArchivedRootProjects)
const favoriteProjects = computed(() => projectStore.favoriteProjects)
const savedFilterProjects = computed(() => projectStore.savedFilterProjects)
</script>

<style lang="scss" scoped>
.logo {
	display: block;

	padding-inline-start: 1rem;
	margin-inline-end: 1rem;
	margin-block-end: 1rem;

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
	inset-block-start: $navbar-height;
	inset-block-end: 0;
	inset-inline-start: 0;
	transform: translateX(-100%);
	inline-size: $navbar-width;
	overflow-y: auto;

	[dir="rtl"] & {
		inset-inline-start: auto;
		inset-inline-end: 0;
		transform: translateX(100%);
	}

	@media screen and (max-width: $tablet) {
		inset-block-start: 0;
		inline-size: 70vw;
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
		padding-inline-start: 2rem;
		display: inline-block;

		.icon {
			padding-block-end: .25rem;
		}
	}
}

.menu + .menu {
	padding-block-start: math.div($navbar-padding, 2);
}
</style>
