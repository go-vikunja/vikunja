<template>
	<div
		class="wiki-page-item"
		:class="{
			'is-active': isActive,
			'is-folder': page.isFolder,
		}"
		:style="{ paddingLeft: `${level * 1.5}rem` }"
	>
		<div class="wiki-page-item-content">
			<BaseButton
				v-if="page.isFolder"
				class="collapse-button"
				@click="toggleCollapsed"
			>
				<Icon
					icon="chevron-down"
					:class="{ 'is-collapsed': isCollapsed }"
				/>
			</BaseButton>
			<span
				v-else
				class="collapse-button-placeholder"
			/>
			
			<BaseButton
				class="page-link"
				:class="{'router-link-exact-active': isActive}"
				@click="handleClick"
				@dblclick.stop="startEditingTitle"
			>
				<Icon :icon="page.isFolder ? 'folder' : 'file-alt'" />
				<input
					v-if="isEditingTitle"
					v-model="editableTitle"
					class="page-title-input"
					@click.stop
					@blur="saveTitle"
					@keydown.enter="saveTitle"
					@keydown.esc="cancelEditTitle"
				>
				<span
					v-else
					class="page-title"
				>{{ page.title }}</span>
			</BaseButton>
			
			<div class="wiki-page-item-actions">
				<BaseButton
					v-if="page.isFolder"
					v-tooltip="$t('wiki.createPageInFolder')"
					class="action-button"
					@click.stop="$emit('createPage', page.id)"
				>
					<Icon icon="plus" />
				</BaseButton>
			</div>
		</div>
		
		<div
			v-if="page.isFolder && !isCollapsed"
			class="wiki-page-children"
		>
			<draggable
				:list="childPages"
				group="wiki-pages"
				:animation="150"
				item-key="id"
				@end="handleDragEnd"
			>
				<template #item="{element: child}">
					<WikiPageItem
						:page="child"
						:project-id="projectId"
						:current-page-id="currentPageId"
						:level="level + 1"
						@pageSelected="$emit('pageSelected', $event)"
						@createPage="$emit('createPage', $event)"
						@createFolder="$emit('createFolder', $event)"
					/>
				</template>
			</draggable>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, nextTick, toRef, watch} from 'vue'
import {useStorage} from '@vueuse/core'
import draggable from 'zhyswan-vuedraggable'
import BaseButton from '@/components/base/BaseButton.vue'
import {useWikiPageStore} from '@/stores/wikiPages'
import type {IWikiPage} from '@/modelTypes/IWikiPage'

// Recursive component reference
import WikiPageItem from './WikiPageItem.vue'

const props = defineProps<{
	page: IWikiPage
	projectId: number
	currentPageId: number | null
	level: number
}>()

const emit = defineEmits<{
	pageSelected: [page: IWikiPage]
	createPage: [parentId: number | null]
	createFolder: [parentId: number | null]
}>()

const wikiPageStore = useWikiPageStore()

const pageRef = toRef(props, 'page')
const isEditingTitle = ref(false)
const editableTitle = ref('')

watch(() => pageRef.value.title, (newTitle) => {
	editableTitle.value = newTitle
}, {immediate: true})

const collapsedState = useStorage<Record<number, boolean>>('wiki-collapsed-folders-v2', {})
const isCollapsed = computed({
	get: () => {
		// Default to expanded (false) for new folders
		return collapsedState.value[pageRef.value.id] ?? false
	},
	set: (value) => {
		collapsedState.value[pageRef.value.id] = value
	},
})

const isActive = computed(() => {
	return !props.page.isFolder && props.currentPageId === props.page.id
})

const childPages = computed(() => {
	if (!props.page.isFolder) return []
	return wikiPageStore.getChildPages(props.projectId, props.page.id)
})

function toggleCollapsed() {
	isCollapsed.value = !isCollapsed.value
}

function handleClick() {
	if (!props.page.isFolder) {
		emit('pageSelected', props.page)
	} else {
		toggleCollapsed()
	}
}

function handleDragEnd(): void {
	// TODO: Implement drag-and-drop reordering
	console.log('Drag ended')
}

function startEditingTitle() {
	isEditingTitle.value = true
	editableTitle.value = props.page.title
	nextTick(() => {
		const input = document.querySelector('.page-title-input') as HTMLInputElement
		if (input) {
			input.focus()
			input.select()
		}
	})
}

function cancelEditTitle() {
	isEditingTitle.value = false
	editableTitle.value = props.page.title
}

async function saveTitle() {
	if (editableTitle.value.trim() === '') {
		cancelEditTitle()
		return
	}
	
	if (editableTitle.value !== props.page.title) {
		const updatedPage = {
			...props.page,
			title: editableTitle.value,
		}
		await wikiPageStore.updateWikiPage(props.projectId, updatedPage)
	}
	
	isEditingTitle.value = false
}
</script>

<style lang="scss" scoped>
.wiki-page-item {
	position: relative;
	
	@media screen and (max-width: $mobile) {
		padding-inline-start: calc(var(--level, 0) * 1rem) !important;
	}
}

.wiki-page-item-content {
	display: flex;
	align-items: center;
	gap: 0.25rem;
	padding: 0.25rem 0.5rem;
	
	@media screen and (max-width: $mobile) {
		padding: 0.35rem 0.25rem;
	}
	
	&:hover {
		background: var(--grey-100);
		
		.wiki-page-item-actions {
			opacity: 1;
		}
	}
}

.collapse-button {
	padding: 0.25rem;
	min-inline-size: auto;
	background: transparent;
	
	:deep(svg),
	:deep(i) {
		transition: transform 0.2s;
	}
}

.collapse-button :deep(.is-collapsed) {
	transform: rotate(-90deg);
}

.collapse-button-placeholder {
	inline-size: 24px;
	min-inline-size: 24px;
	
	@media screen and (max-width: $mobile) {
		inline-size: 20px;
		min-inline-size: 20px;
	}
}

.page-link {
	flex: 1;
	justify-content: flex-start;
	gap: 0.5rem;
	padding: .25rem .5rem;
	display: block;
	border-radius: $radius;
	font-size: 0.9rem;
	text-align: start;
	color: var(--text);
	background: transparent;
	transition: all 100ms;
	
	@media screen and (max-width: $mobile) {
		padding: .25rem 0.25rem;
		font-size: 0.85rem;
		gap: 0.35rem;
	}
	
	&:hover {
		color: var(--switch-view-color);
		background: var(--primary);
	}
	
	&.router-link-exact-active {
		color: var(--switch-view-color);
		background: var(--primary);
		font-weight: bold;
		box-shadow: var(--shadow-xs);
	}
}

.page-title {
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.title-input {
	flex: 1;
	padding: 0.35rem 0.5rem;
	font-size: 0.9rem;
	border: 1px solid var(--primary);
	border-radius: 4px;
	background: var(--white);
	color: var(--text);
	
	@media screen and (max-width: $mobile) {
		padding: 0.25rem 0.35rem;
		font-size: 0.85rem;
	}
	
	&:focus {
		outline: none;
		box-shadow: 0 0 0 2px var(--primary-light);
	}
}

.wiki-page-item-actions {
	display: flex;
	gap: 0.25rem;
	opacity: 0;
	transition: opacity 0.2s;
	
	@media screen and (max-width: $mobile) {
		opacity: 1;
	}
}

.action-button {
	padding: 0.25rem 0.5rem;
	min-inline-size: auto;
	
	@media screen and (max-width: $mobile) {
		padding: 0.25rem 0.35rem;
	}
}

.is-folder {
	.page-link {
		font-weight: 600;
	}
}
</style>
