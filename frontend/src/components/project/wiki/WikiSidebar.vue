<template>
	<div class="wiki-sidebar-container">
		<div class="wiki-sidebar-header">
			<h3 class="wiki-sidebar-title">
				{{ $t('wiki.title') }}
			</h3>
			<div class="wiki-sidebar-actions">
				<BaseButton
					v-tooltip="$t('wiki.createPage')"
					@click="$emit('createPage', null)"
				>
					<Icon icon="file-alt" />
				</BaseButton>
				<BaseButton
					v-tooltip="$t('wiki.createFolder')"
					@click="$emit('createFolder', null)"
				>
					<Icon icon="folder" />
				</BaseButton>
			</div>
		</div>
		
		<draggable
			:list="rootPages"
			class="wiki-page-tree"
			group="wiki-pages"
			:animation="150"
			item-key="id"
			@end="handleDragEnd"
		>
			<template #item="{element: page}">
				<WikiPageItem
					:page="page"
					:project-id="projectId"
					:current-page-id="currentPageId"
					:level="0"
					@pageSelected="$emit('pageSelected', $event)"
					@createPage="$emit('createPage', $event)"
					@createFolder="$emit('createFolder', $event)"
				/>
			</template>
		</draggable>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import draggable from 'zhyswan-vuedraggable'
import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/misc/Icon'
import WikiPageItem from './WikiPageItem.vue'
import {useWikiPageStore} from '@/stores/wikiPages'

const props = defineProps<{
	projectId: number
	currentPageId: number | null
}>()

defineEmits<{
	pageSelected: [page: IWikiPage]
	createPage: [parentId: number | null]
	createFolder: [parentId: number | null]
}>()

const wikiPageStore = useWikiPageStore()

const rootPages = computed(() => {
	return wikiPageStore.getRootPagesForProject(props.projectId)
})

function handleDragEnd(event: any) {
	// TODO: Implement drag-and-drop reordering at root level
	console.log('Root level drag ended', event)
}
</script>

<style lang="scss" scoped>
.wiki-sidebar-container {
	display: flex;
	flex-direction: column;
	height: 100%;
}

.wiki-sidebar-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 1rem;
	border-bottom: 1px solid var(--grey-200);
}

.wiki-sidebar-title {
	font-size: 1.1rem;
	font-weight: 600;
	margin: 0;
}

.wiki-sidebar-actions {
	display: flex;
	gap: 0.5rem;
	
	button {
		background: var(--grey-100);
		color: var(--grey-600);
		padding: 0.5rem;
		border-radius: 4px;
		
		&:hover {
			background: var(--grey-200);
			color: var(--grey-800);
		}
	}
}

.wiki-page-tree {
	flex: 1;
	overflow-y: auto;
	padding: 0.5rem 0;
}
</style>
