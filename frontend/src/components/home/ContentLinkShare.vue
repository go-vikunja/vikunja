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
			<h1
				:class="{'m-0': !logoVisible}"
				:style="{ 'opacity': currentProject?.title === '' ? '0': '1' }"
				class="title"
			>
				{{ currentProject?.title === '' ? $t('misc.loading') : currentProject?.title }}
			</h1>
			<div class="box has-text-start view">
				<RouterView />
				<PoweredByLink utm-medium="link_share" />
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'

import {useBaseStore} from '@/stores/base'
import {useRoute} from 'vue-router'

import Logo from '@/components/home/Logo.vue'
import PoweredByLink from './PoweredByLink.vue'
import {useProjectStore} from '@/stores/projects'
import {useLabelStore} from '@/stores/labels'
import {PROJECT_VIEW_KINDS} from '@/modelTypes/IProjectView'

const baseStore = useBaseStore()
const currentProject = computed(() => baseStore.currentProject)
const background = computed(() => baseStore.background)
const logoVisible = computed(() => baseStore.logoVisible)

const projectStore = useProjectStore()
projectStore.loadAllProjects()

const labelStore = useLabelStore()
labelStore.loadAllLabels()

const route = useRoute()
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
