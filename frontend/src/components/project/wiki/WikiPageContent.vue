<template>
	<div class="wiki-page-content">
		<div class="wiki-page-header">
			<input
				v-if="isEditingTitle"
				ref="titleInputRef"
				v-model="editableTitle"
				class="wiki-page-title-input"
				@blur="saveTitle"
				@keydown.enter="saveTitle"
			>
			<h1
				v-else
				class="wiki-page-title"
				:title="$t('wiki.doubleclickToEdit')"
				@dblclick="startEditingTitle"
			>
				{{ editableTitle }}
			</h1>
		</div>
		
		<TipTap
			:model-value="editableContent"
			:show-save="true"
			:is-edit-enabled="true"
			:start-in-edit-mode="!editableContent || editableContent.trim() === '' || editableContent === '<p></p>'"
			:placeholder="$t('wiki.contentPlaceholder')"
			@update:modelValue="updateContent"
			@save="handleSave"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, watch, nextTick, toRef} from 'vue'
import TipTap from '@/components/input/editor/TipTap.vue'
import type {IWikiPage} from '@/modelTypes/IWikiPage'

const props = defineProps<{
	page: IWikiPage
}>()

const emit = defineEmits<{
	save: [page: IWikiPage]
}>()

const pageTitle = ref(props.page.title)
const pageContent = ref(props.page.content)

const page = toRef(props, 'page')
const editableTitle = ref('')
const editableContent = ref('')
const isEditingTitle = ref(false)
const titleInputRef = ref<HTMLInputElement>()

// Initialize values from props
watch(() => page.value, (newPage) => {
	editableTitle.value = newPage.title
	editableContent.value = newPage.content
}, { immediate: true })

function startEditingTitle() {
	isEditingTitle.value = true
	nextTick(() => {
		titleInputRef.value?.focus()
		titleInputRef.value?.select()
	})
}

function saveTitle() {
	isEditingTitle.value = false
	if (editableTitle.value !== page.value.title) {
		handleSave()
	}
}

function updateContent(newContent: string) {
	editableContent.value = newContent
}

function handleSave() {
	const updatedPage: IWikiPage = {
		...page.value,
		title: editableTitle.value,
		content: editableContent.value,
	}
	emit('save', updatedPage)
}
</script>

<style lang="scss" scoped>
.wiki-page-content {
	background: var(--white);
	padding: 1rem;
}

.wiki-page-header {
	margin-bottom: 1rem;
	padding-bottom: 1rem;
	border-bottom: 1px solid var(--grey-200);
}

.wiki-page-title {
	font-size: 1.75rem;
	font-weight: 700;
	margin: 0;
	cursor: pointer;
	transition: color $transition;
	
	&:hover {
		color: var(--primary);
	}
}

.wiki-page-title-input {
	width: 100%;
	font-size: 1.75rem;
	font-weight: 700;
	border: none;
	border-bottom: 2px solid var(--primary);
	padding: 0.25rem 0;
	outline: none;
	background: transparent;
	
	&:focus {
		border-bottom-color: var(--primary);
	}
}
</style>
