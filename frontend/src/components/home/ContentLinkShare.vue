<template>
	<div
		:class="{
			'has-background': background,
			'link-share-is-fullwidth': isFullWidth,
		}"
		:style="{'background-image': `url(${background})`}"
		class="link-share-container"
	>
		<div class="has-text-centered link-share-view">
			<Logo
				v-if="logoVisible"
				class="logo"
			/>
			<Message
				v-if="projectLoadError"
				variant="danger"
				class="mbe-4"
			>
				{{ $t('sharing.projectLoadError') }}
				<BaseButton
					variant="secondary"
					class="mls-2"
					@click="retryProjectLoad"
				>
					{{ $t('sharing.retry') }}
				</BaseButton>
			</Message>
			<BaseButton
				v-if="!projectLoadError && currentProject && getProjectRoute()"
				:to="getProjectRoute()!"
				variant="text"
				class="project-title-button"
				:class="{'m-0': !logoVisible}"
			>
				<h1 class="title clickable-title">
					{{ currentProject?.title === '' ? $t('misc.loading') : currentProject?.title }}
				</h1>
			</BaseButton>
			<h1
				v-else-if="!projectLoadError"
				:class="{'m-0': !logoVisible}"
				class="title"
			>
				{{ $t('misc.loading') }}
			</h1>
			<div class="box has-text-start view">
				<RouterView />
				<PoweredByLink utm-medium="link_share" />
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {computed, ref, watch, onMounted} from 'vue'
import {useRoute} from 'vue-router'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useLabelStore} from '@/stores/labels'
import {useAuthStore} from '@/stores/auth'

import Logo from '@/components/home/Logo.vue'
import PoweredByLink from './PoweredByLink.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import Message from '@/components/misc/Message.vue'
import {PROJECT_VIEW_KINDS} from '@/modelTypes/IProjectView'
import {getRouteParamAsNumber} from '@/helpers/utils'

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const authStore = useAuthStore()
const route = useRoute()

const currentProject = computed(() => baseStore.currentProject)
const background = computed(() => baseStore.background)
const logoVisible = computed(() => baseStore.logoVisible)
const projectLoadError = ref(false)

projectStore.loadAllProjects()

const labelStore = useLabelStore()
labelStore.loadAllLabels()

// Ensure project is loaded for link share
async function ensureProjectLoaded() {
	if (!authStore.authLinkShare || !route.params.projectId) {
		return
	}
	
	try {
		projectLoadError.value = false
		
		// Load project if not already loaded
		const projectId = getRouteParamAsNumber(route.params.projectId)
		if (projectId && (!currentProject.value || currentProject.value.id !== projectId)) {
			await projectStore.loadProject(projectId)
		}
	} catch (e) {
		console.error('Failed to load project for link share:', e)
		projectLoadError.value = true
	}
}

async function retryProjectLoad() {
	await ensureProjectLoaded()
}

// Watch for route changes and ensure project is loaded
watch(() => route.params.projectId, ensureProjectLoaded, { immediate: true })

onMounted(ensureProjectLoaded)

function getProjectRoute() {
	if (!currentProject.value) return null
	
	const hash = route.hash // Preserve link share hash
	
	// Default to the first available view or list view
	const projectId = currentProject.value.id
	const firstView = projectStore.projects[projectId]?.views?.[0]
	
	if (firstView) {
		return {
			name: 'project.view',
			params: { projectId, viewId: firstView.id },
			hash,
		}
	}
	
	return {
		name: 'project.index', 
		params: { projectId },
		hash,
	}
}

const isFullWidth = computed(() => {
	const viewId = route.params?.viewId ?? null
	const projectId = route.params?.projectId ?? null
	if (!viewId || !projectId) {
		return false
	}

	const view = projectStore.projects[Number(projectId)]?.views.find(v => v.id === Number(viewId))

	return view?.viewKind === PROJECT_VIEW_KINDS.KANBAN ||
		view?.viewKind === PROJECT_VIEW_KINDS.GANTT
})
</script>

<style lang="scss" scoped>
.link-share-container.has-background .view {
	background-color: transparent;
	border: none;
}

.logo {
	max-inline-size: 300px;
	inline-size: 90%;
	margin: 1rem auto 2rem;
	block-size: 100px;
}

.title {
	text-shadow: 0 0 1rem var(--white);
}

.project-title-button {
	background: none !important;
	border: none !important;
	padding: 0 !important;
	text-decoration: none !important;
	
	&:hover .clickable-title {
		opacity: 0.8;
		cursor: pointer;
	}
}

.clickable-title {
	text-shadow: 0 0 1rem var(--white);
	margin: 0;
	
	&:hover {
		text-decoration: underline;
	}
}

.link-share-view {
	inline-size: 100%;
	max-inline-size: $desktop;
	margin: 0 auto;
}

.link-share-container.link-share-is-fullwidth {
	.link-share-view {
		max-inline-size: 100vw;
	}
}

.link-share-container:not(.has-background) {
	:deep(.loader-container, .gantt-chart-container > .card) {
		box-shadow: none !important;
		border: none;

		.task-add {
			padding: 1rem 0 0;
		}
	}
}
</style>
