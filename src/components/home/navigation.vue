<template>
	<aside :class="{'is-active': menuActive}" class="namespace-container">
		<nav class="menu top-menu">
			<router-link :to="{name: 'home'}" class="logo">
				<Logo width="164" height="48"/>
			</router-link>
			<ul class="menu-list">
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
		
		<nav>
			<template v-for="(p, pk) in projects" :key="p.id">
				{{ p. title }}<br/>
			</template>
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
<!--				<draggable-->
<!--						v-if="projectsVisible[n.id] ?? true"-->
<!--						v-bind="dragOptions"-->
<!--						:modelValue="activeProjects[nk]"-->
<!--						@update:modelValue="(projects) => updateActiveProjects(n, projects)"-->
<!--						group="namespace-lists"-->
<!--						@start="() => drag = true"-->
<!--						@end="saveListPosition"-->
<!--						handle=".handle"-->
<!--						:disabled="n.id < 0 || undefined"-->
<!--						tag="ul"-->
<!--						item-key="id"-->
<!--						:data-namespace-id="n.id"-->
<!--						:data-namespace-index="nk"-->
<!--						:component-data="{-->
<!--							type: 'transition-group',-->
<!--							name: !drag ? 'flip-list' : null,-->
<!--							class: [-->
<!--								'menu-list can-be-hidden',-->
<!--								{ 'dragging-disabled': n.id < 0 }-->
<!--							]-->
<!--						}"-->
<!--				>-->
<!--					<template #item="{element: l}">-->
<!--						<li-->
<!--							class="list-menu loader-container is-loading-small"-->
<!--							:class="{'is-loading': projectUpdating[l.id]}"-->
<!--						>-->
<!--							<BaseButton-->
<!--								:to="{ name: 'project.index', params: { projectId: l.id} }"-->
<!--								class="list-menu-link"-->
<!--								:class="{'router-link-exact-active': currentProject.id === l.id}"-->
<!--							>-->
<!--								<span class="icon menu-item-icon handle">-->
<!--									<icon icon="grip-lines"/>-->
<!--								</span>-->
<!--								<ColorBubble-->
<!--										v-if="l.hexColor !== ''"-->
<!--										:color="l.hexColor"-->
<!--										class="mr-1"-->
<!--								/>-->
<!--								<span class="list-menu-title">{{ getProjectTitle(l) }}</span>-->
<!--							</BaseButton>-->
<!--							<BaseButton-->
<!--								v-if="l.id > 0"-->
<!--								class="favorite"-->
<!--								:class="{'is-favorite': l.isFavorite}"-->
<!--								@click="projectStore.toggleProjectFavorite(l)"-->
<!--							>-->
<!--								<icon :icon="l.isFavorite ? 'star' : ['far', 'star']"/>-->
<!--							</BaseButton>-->
<!--							<ProjectSettingsDropdown class="menu-list-dropdown" :project="l" v-if="l.id > 0">-->
<!--								<template #trigger="{toggleOpen}">-->
<!--									<BaseButton class="menu-list-dropdown-trigger" @click="toggleOpen">-->
<!--										<icon icon="ellipsis-h" class="icon"/>-->
<!--									</BaseButton>-->
<!--								</template>-->
<!--							</ProjectSettingsDropdown>-->
<!--							<span class="list-setting-spacer" v-else></span>-->
<!--						</li>-->
<!--					</template>-->
<!--				</draggable>-->
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
import NamespaceSettingsDropdown from '@/components/namespace/namespace-settings-dropdown.vue'
import PoweredByLink from '@/components/home/PoweredByLink.vue'
import Logo from '@/components/home/Logo.vue'

import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import {getNamespaceTitle} from '@/helpers/getNamespaceTitle'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import type {IProject} from '@/modelTypes/IProject'
import type {INamespace} from '@/modelTypes/INamespace'
import ColorBubble from '@/components/misc/colorBubble.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useNamespaceStore} from '@/stores/namespaces'

const drag = ref(false)
const dragOptions = {
	animation: 100,
	ghostClass: 'ghost',
}

const baseStore = useBaseStore()
const namespaceStore = useNamespaceStore()
const currentProject = computed(() => baseStore.currentProject)
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
	return namespaces.value.map((namespace) => getNamespaceTitle(namespace))
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

const projects = computed(() => projectStore.projects)

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

const projectUpdating = ref<{ [id: INamespace['id']]: boolean }>({})

async function saveProjectPosition(e: SortableEvent) {
	if (!e.newIndex && e.newIndex !== 0) return

	const namespaceId = parseInt(e.to.dataset.namespaceId as string)
	const newNamespaceIndex = parseInt(e.to.dataset.namespaceIndex as string)

	const projectsActive = activeProjects.value[newNamespaceIndex]
	// If the project was dragged to the last position, Safari will report e.newIndex as the size of the projectsActive
	// array instead of using the position. Because the index is wrong in that case, dragging the project will fail.
	// To work around that we're explicitly checking that case here and decrease the index.
	const newIndex = e.newIndex === projectsActive.length ? e.newIndex - 1 : e.newIndex

	const project = projectsActive[newIndex]
	const projectBefore = projectsActive[newIndex - 1] ?? null
	const projectAfter = projectsActive[newIndex + 1] ?? null
	projectUpdating.value[project.id] = true

	const position = calculateItemPosition(
		projectBefore !== null ? projectBefore.position : null,
		projectAfter !== null ? projectAfter.position : null,
	)

	try {
		// create a copy of the project in order to not violate pinia manipulation
		await projectStore.updateProject({
			...project,
			position,
			namespaceId,
		})
	} finally {
		projectUpdating.value[project.id] = false
	}
}
</script>

<style lang="scss" scoped>
$navbar-padding: 2rem;
$vikunja-nav-background: var(--site-background);
$vikunja-nav-color: var(--grey-700);
$vikunja-nav-selected-width: 0.4rem;

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
	background: $vikunja-nav-background;
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

// these are general menu styles
// should be in own components
.menu {
	.menu-label,
	.menu-list .list-menu-link,
	.menu-list a {
		display: flex;
		align-items: center;
		justify-content: space-between;
		cursor: pointer;

		.color-bubble {
			height: 12px;
			flex: 0 0 12px;
		}
	}

	.menu-list {
		li {
			height: 44px;
			display: flex;
			align-items: center;

			&:hover {
				background: var(--white);
			}

			.menu-list-dropdown {
				opacity: 1;
				transition: $transition;
			}

			@media(hover: hover) and (pointer: fine) {
				.menu-list-dropdown {
					opacity: 0;
				}

				&:hover .menu-list-dropdown {
					opacity: 1;
				}
			}

		}

		.menu-item-icon {
			color: var(--grey-400);
		}

		.menu-list-dropdown-trigger {
			display: flex;
			padding: 0.5rem;
		}

		.flip-list-move {
			transition: transform $transition-duration;
		}

		.ghost {
			background: var(--grey-200);

			* {
				opacity: 0;
			}
		}

		a:hover {
			background: transparent;
		}

		.list-menu-link,
		li > a {
			color: $vikunja-nav-color;
			padding: 0.75rem .5rem 0.75rem ($navbar-padding * 1.5 - 1.75rem);
			transition: all 0.2s ease;

			border-radius: 0;
			white-space: nowrap;
			text-overflow: ellipsis;
			overflow: hidden;
			width: 100%;
			border-left: $vikunja-nav-selected-width solid transparent;

			&:hover {
				border-left: $vikunja-nav-selected-width solid var(--primary);
			}

			&.router-link-exact-active {
				color: var(--primary);
				border-left: $vikunja-nav-selected-width solid var(--primary);
			}

			.icon {
				height: 1rem;
				vertical-align: middle;
				padding-right: 0.5rem;
			}

			&.router-link-exact-active .icon:not(.handle) {
				color: var(--primary);
			}

			.handle {
				opacity: 0;
				transition: opacity $transition;
				margin-right: .25rem;
			}

			&:hover .handle {
				opacity: 1;
			}
		}
		&:not(.dragging-disabled) .handle {
			cursor: grab;
		}
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

.namespace-title {
	display: flex;
	align-items: center;
	justify-content: space-between;
	color: $vikunja-nav-color;
	padding: 0 .25rem;

	.toggle-lists-icon {
		svg {
			transition: all $transition;
			transform: rotate(90deg);
			opacity: 1;
		}

		&.active svg {
			transform: rotate(0deg);
			opacity: 0;
		}
	}

	&:hover .toggle-lists-icon svg {
		opacity: 1;
	}

	&:not(.has-menu) .toggle-lists-icon {
		padding-right: 1rem;
	}
}

.list-setting-spacer {
	width: 2.5rem;
	flex-shrink: 0;
}

.namespaces-list.loader-container.is-loading {
	min-height: calc(100vh - #{$navbar-height + 1.5rem + 1rem + 1.5rem});
}
</style>
