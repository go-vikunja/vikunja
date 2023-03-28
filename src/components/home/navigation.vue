<template>
	<aside :class="{'is-active': menuActive}" class="menu-container">
		<nav class="menu top-menu">
			<router-link :to="{name: 'home'}" class="logo">
				<Logo width="164" height="48"/>
			</router-link>
			<ul class="menu-list other-menu-items">
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
					<router-link :to="{ name: 'projects.index'}" v-shortcut="'g n'">
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
			</ul>
		</nav>

		<nav class="menu" v-if="favoriteProjects">
			<ProjectsNavigation v-model="favoriteProjects" :allow-drag="false"/>
		</nav>

		<nav class="menu">
			<ProjectsNavigation v-model="projects" :allow-drag="true"/>
		</nav>

		<PoweredByLink/>
	</aside>
</template>

<script setup lang="ts">
import {computed, onBeforeMount} from 'vue'

import PoweredByLink from '@/components/home/PoweredByLink.vue'
import Logo from '@/components/home/Logo.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const menuActive = computed(() => baseStore.menuActive)

// FIXME: async action will be unfinished when component mounts
onBeforeMount(async () => {
	await projectStore.loadProjects()
})

const projects = computed(() => Object.values(projectStore.projects)
	.filter(p => p.parentProjectId === 0 && !p.isArchived)
	.sort((a, b) => a.position < b.position ? -1 : 1))
const favoriteProjects = computed(() => Object.values(projectStore.projects)
	.filter(p => !p.isArchived && p.isFavorite)
	.map(p => ({
		...p,
		childProjects: [],
	}))
	.sort((a, b) => a.position < b.position ? -1 : 1))
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

.top-menu {
	margin-top: math.div($navbar-padding, 2);

	.menu-list {
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
}

.menu {
	padding-top: math.div($navbar-padding, 2);
}
</style>
