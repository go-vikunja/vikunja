<template>
	<aside :class="{'is-active': menuActive}" class="menu-container">
		<nav class="menu top-menu">
			<router-link :to="{name: 'home'}" class="logo">
				<Logo width="164" height="48"/>
			</router-link>
			<menu class="menu-list other-menu-items">
				<li>
					<router-link :to="{ name: 'home'}" v-shortcut="'g o'">
						<span class="menu-item-icon icon">
							<icon icon="calendar"/>
						</span>
						{{ $t('navigation.overview') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'tasks.range'}" v-shortcut="'g u'">
						<span class="menu-item-icon icon">
							<icon :icon="['far', 'calendar-alt']"/>
						</span>
						{{ $t('navigation.upcoming') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'projects.index'}" v-shortcut="'g p'">
						<span class="menu-item-icon icon">
							<icon icon="layer-group"/>
						</span>
						{{ $t('project.projects') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'labels.index'}" v-shortcut="'g a'">
						<span class="menu-item-icon icon">
							<icon icon="tags"/>
						</span>
						{{ $t('label.title') }}
					</router-link>
				</li>
				<li>
					<router-link :to="{ name: 'teams.index'}" v-shortcut="'g m'">
						<span class="menu-item-icon icon">
							<icon icon="users"/>
						</span>
						{{ $t('team.title') }}
					</router-link>
				</li>
			</menu>
		</nav>

		<Loading
			v-if="projectsLoading"
			variant="small"
		/>
		<template v-else>
			<nav class="menu" v-if="favoriteProjects">
				<ProjectsNavigation :model-value="favoriteProjects" :can-edit-order="false" :can-collapse="false"/>
			</nav>

			<nav class="menu">
				<ProjectsNavigation
					:model-value="projects"
					:can-edit-order="true"
					:can-collapse="true"
					:level="1"
				/>
			</nav>
		</template>

		<PoweredByLink/>
	</aside>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import PoweredByLink from '@/components/home/PoweredByLink.vue'
import Logo from '@/components/home/Logo.vue'
import Loading from '@/components/misc/loading.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const menuActive = computed(() => baseStore.menuActive)
const projectsLoading = computed(() => projectStore.isLoading)

const projects = computed(() => projectStore.notArchivedRootProjects
	.sort((a, b) => a.position - b.position))
const favoriteProjects = computed(() => projectStore.favoriteProjects
	.sort((a, b) => a.position - b.position))
</script>

<style lang="scss" scoped>
.logo {
	display: block;

	padding-left: 1rem;
	margin-right: 1rem;
	margin-bottom: 1rem;

	@media screen and (min-width: $tablet) {
		display: none;
	}
}

.menu-container {
	background: var(--site-background);
	color: $vikunja-nav-color;
	padding: 0 0 1rem;
	transition: transform $transition-duration ease-in;
	position: fixed;
	top: $navbar-height;
	bottom: 0;
	left: 0;
	transform: translateX(-100%);
	overflow-x: auto;
	width: $navbar-width;

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

.menu + .menu {
	padding-top: math.div($navbar-padding, 2);
}
</style>
