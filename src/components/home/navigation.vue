<template>
	<aside :class="{'is-active': menuActive}" class="namespace-container">
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

		<nav class="menu namespaces-lists">
			<ProjectsNavigation :projects="projects"/>
		</nav>

		<!--		<nav class="menu namespaces-lists loader-container is-loading-small" :class="{'is-loading': loading}">-->
		<!--			<template v-for="(n, nk) in namespaces" :key="n.id">-->
		<!--				<div class="namespace-title" :class="{'has-menu': n.id > 0}">-->
		<!--					<BaseButton-->
		<!--						@click="toggleProjects(n.id)"-->
		<!--						class="menu-label"-->
		<!--						v-tooltip="namespaceTitles[nk]"-->
		<!--					>-->
		<!--						<ColorBubble-->
		<!--								v-if="n.hexColor !== ''"-->
		<!--								:color="n.hexColor"-->
		<!--								class="mr-1"-->
		<!--						/>-->
		<!--						<span class="name">{{ namespaceTitles[nk] }}</span>-->
		<!--						<div-->
		<!--								class="icon menu-item-icon is-small toggle-lists-icon pl-2"-->
		<!--								:class="{'active': typeof projectsVisible[n.id] !== 'undefined' ? projectsVisible[n.id] : true}"-->
		<!--						>-->
		<!--							<icon icon="chevron-down"/>-->
		<!--						</div>-->
		<!--						<span class="count" :class="{'ml-2 mr-0': n.id > 0}">-->
		<!--							({{ namespaceProjectsCount[nk] }})-->
		<!--						</span>-->
		<!--					</BaseButton>-->
		<!--					<namespace-settings-dropdown class="menu-list-dropdown" :namespace="n" v-if="n.id > 0"/>-->
		<!--				</div>-->
		<!--				&lt;!&ndash;-->
		<!--					NOTE: a v-model / computed setter is not possible, since the updateActiveProjects function-->
		<!--					triggered by the change needs to have access to the current namespace-->
		<!--				&ndash;&gt;-->

		<!--			</template>-->
		<!--		</nav>-->
		<PoweredByLink/>
	</aside>
</template>

<script setup lang="ts">
import {ref, computed, onBeforeMount} from 'vue'
import draggable from 'zhyswan-vuedraggable'
import type {SortableEvent} from 'sortablejs'

import BaseButton from '@/components/base/BaseButton.vue'
import ProjectSettingsDropdown from '@/components/project/project-settings-dropdown.vue'
import PoweredByLink from '@/components/home/PoweredByLink.vue'
import Logo from '@/components/home/Logo.vue'

import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import type {IProject} from '@/modelTypes/IProject'
import type {INamespace} from '@/modelTypes/INamespace'
import ColorBubble from '@/components/misc/colorBubble.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'


const baseStore = useBaseStore()
const menuActive = computed(() => baseStore.menuActive)
const loading = computed(() => namespaceStore.isLoading)


const namespaces = computed(() => {
	return namespaceStore.namespaces.filter(n => !n.isArchived)
})
const activeProjects = computed(() => {
	return namespaces.value.map(({projects}) => {
		return projects?.filter(item => {
			return typeof item !== 'undefined' && !item.isArchived
		})
	})
})

const namespaceTitles = computed(() => {
	return []
})

const namespaceProjectsCount = computed(() => {
	return namespaces.value.map((_, index) => activeProjects.value[index]?.length ?? 0)
})

const projectStore = useProjectStore()

function toggleProjects(namespaceId: INamespace['id']) {
	projectsVisible.value[namespaceId] = !projectsVisible.value[namespaceId]
}

const projectsVisible = ref<{ [id: INamespace['id']]: boolean }>({})
// FIXME: async action will be unfinished when component mounts
onBeforeMount(async () => {
	await projectStore.loadProjects()
})

const projects = computed(() => Object.values(projectStore.projects).sort((a, b) => a.position < b.position ? -1 : 1))

function updateActiveProjects(namespace: INamespace, activeProjects: IProject[]) {
	// This is a bit hacky: since we do have to filter out the archived items from the list
	// for vue draggable updating it is not as simple as replacing it.
	// To work around this, we merge the active projects with the archived ones. Doing so breaks the order
	// because now all archived projects are sorted after the active ones. This is fine because they are sorted
	// later when showing them anyway, and it makes the merging happening here a lot easier.
	const projects = [
		...activeProjects,
		...namespace.projects.filter(l => l.isArchived),
	]

	namespaceStore.setNamespaceById({
		...namespace,
		projects,
	})
}


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

.namespace-container {
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

.namespaces-lists {
	padding-top: math.div($navbar-padding, 2);

	.menu-label {
		font-size: 1rem;
		font-weight: 700;
		font-weight: bold;
		font-family: $vikunja-font;
		color: $vikunja-nav-color;
		font-weight: 600;
		min-height: 2.5rem;
		padding-top: 0;
		padding-left: $navbar-padding;

		overflow: hidden;
		margin-bottom: 0;
		flex: 1 1 auto;

		.name {
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
			margin-right: auto;
		}

		.count {
			color: var(--grey-500);
			margin-right: .5rem;
			// align brackets with number
			font-feature-settings: "case";
		}
	}

	.favorite {
		margin-left: .25rem;
		transition: opacity $transition, color $transition;
		opacity: 1;

		&.is-favorite {
			color: var(--warning);
			opacity: 1;
		}
	}


	@media(hover: hover) and (pointer: fine) {
		.list-menu .favorite {
			opacity: 0;
		}

		.list-menu:hover .favorite,
		.favorite.is-favorite {
			opacity: 1;
		}
	}

	.list-menu-title {
		overflow: hidden;
		text-overflow: ellipsis;
		width: 100%;
	}

	.color-bubble {
		width: 14px;
		height: 14px;
		flex-basis: auto;
	}

	.is-archived {
		min-width: 85px;
	}
}
</style>
