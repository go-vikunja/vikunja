<template>
	<ProjectWrapper
		class="wiki-view"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id="viewId"
	>
		<div class="loader-container">
			<div class="wiki-layout">
				<WikiSidebar
					:project-id="projectId"
					:current-page-id="currentPageId"
					class="wiki-sidebar"
					@pageSelected="handlePageSelected"
					@createPage="handleCreatePage"
					@createFolder="handleCreateFolder"
				/>
				<div class="wiki-content">
					<WikiBreadcrumbs
						v-if="currentPage"
						:project-id="projectId"
						:page="currentPage"
						class="wiki-breadcrumbs"
					/>
					<WikiPageContent
						v-if="currentPage"
						:key="currentPage.id"
						:page="currentPage"
						@save="handleSave"
					/>
					<div
						v-else
						class="wiki-empty"
					>
						<p class="has-text-grey">
							{{ $t('wiki.selectOrCreate') }}
						</p>
						<BaseButton
							variant="primary"
							@click="handleCreatePage(null)"
						>
							{{ $t('wiki.createFirstPage') }}
						</BaseButton>
					</div>
				</div>
			</div>
		</div>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, onMounted} from 'vue'
import {useRoute, useRouter} from 'vue-router'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import WikiSidebar from './WikiSidebar.vue'
import WikiBreadcrumbs from './WikiBreadcrumbs.vue'
import WikiPageContent from './WikiPageContent.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import {useWikiPageStore} from '@/stores/wikiPages'
import type {IWikiPage} from '@/modelTypes/IWikiPage'

const props = defineProps<{
	projectId: number
	viewId: number
	isLoadingProject: boolean
}>()

const route = useRoute()
const router = useRouter()
const wikiPageStore = useWikiPageStore()

const currentPageId = computed(() => {
	const pageId = route.query.pageId
	return pageId ? Number(pageId) : null
})

const currentPage = computed(() => {
	if (!currentPageId.value) return null
	return wikiPageStore.getPageById(props.projectId, currentPageId.value)
})

onMounted(async () => {
	await wikiPageStore.loadWikiPagesForProject(props.projectId)
})

function handlePageSelected(page: IWikiPage) {
	if (page.isFolder) return
	
	router.push({
		query: {
			pageId: page.id,
		},
	})
}

async function handleCreatePage(parentId: number | null) {
	const newPage: Partial<IWikiPage> = {
		projectId: props.projectId,
		parentId: parentId,
		title: 'New Page',
		content: '',
		isFolder: false,
	}
	
	const created = await wikiPageStore.createWikiPage(props.projectId, newPage as IWikiPage)
	router.push({
		query: {
			pageId: created.id,
		},
	})
}

async function handleCreateFolder(parentId: number | null) {
	const newFolder: Partial<IWikiPage> = {
		projectId: props.projectId,
		parentId: parentId,
		title: 'New Folder',
		content: '',
		isFolder: true,
	}
	
	await wikiPageStore.createWikiPage(props.projectId, newFolder as IWikiPage)
	// Reload pages to show the new folder
	await wikiPageStore.loadWikiPagesForProject(props.projectId)
}

async function handleSave(page: IWikiPage) {
	await wikiPageStore.updateWikiPage(props.projectId, page)
}
</script>

<style lang="scss" scoped>
.wiki-view {
	:deep(.loader-container) {
		height: 100%;
	}
}

.wiki-layout {
	display: flex;
	min-height: calc(100vh - 140px);
	gap: 0;
	background: var(--white);
	border-radius: $radius;
	box-shadow: var(--shadow-sm);
	overflow: hidden;
}

.wiki-sidebar {
	width: 300px;
	min-width: 300px;
	border-right: 1px solid var(--grey-200);
	overflow-y: auto;
	background: var(--white);
}

.wiki-content {
	flex: 1;
	overflow-y: auto;
	padding: 1rem;
}

.wiki-breadcrumbs {
	margin-bottom: 1rem;
}

.wiki-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 1rem;
	padding: 4rem 2rem;
	text-align: center;
	min-height: 400px;
}
</style>
