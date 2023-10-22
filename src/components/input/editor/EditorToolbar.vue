<template>
	<div class="editor-toolbar">
		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleHeading({ level: 1 }).run()"
				:class="{ 'is-active': editor.isActive('heading', { level: 1 }) }"
				v-tooltip="$t('input.editor.heading1')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-header']"/>
					<span class="icon__lower-text">1</span>
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
				:class="{ 'is-active': editor.isActive('heading', { level: 2 }) }"
				v-tooltip="$t('input.editor.heading2')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-header']"/>
					<span class="icon__lower-text">2</span>
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleHeading({ level: 3 }).run()"
				:class="{ 'is-active': editor.isActive('heading', { level: 3 }) }"
				v-tooltip="$t('input.editor.heading3')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-header']"/>
					<span class="icon__lower-text">3</span>
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleBold().run()"
				:class="{ 'is-active': editor.isActive('bold') }"
				v-tooltip="$t('input.editor.bold')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-bold']"/>
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleItalic().run()"
				:class="{ 'is-active': editor.isActive('italic') }"
				v-tooltip="$t('input.editor.italic')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-italic']"/>
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleUnderline().run()"
				:class="{ 'is-active': editor.isActive('underline') }"
				v-tooltip="$t('input.editor.underline')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-underline']"/>
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleStrike().run()"
				:class="{ 'is-active': editor.isActive('strike') }"
				v-tooltip="$t('input.editor.strikethrough')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-strikethrough']"/>
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleCodeBlock().run()"
				:class="{ 'is-active': editor.isActive('codeBlock') }"
				v-tooltip="$t('input.editor.code')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-code']"/>
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleBlockquote().run()"
				:class="{ 'is-active': editor.isActive('blockquote') }"
				v-tooltip="$t('input.editor.quote')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-quote-right']"/>
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleBulletList().run()"
				:class="{ 'is-active': editor.isActive('bulletList') }"
				v-tooltip="$t('input.editor.bulletList')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-list-ol']"/>
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleOrderedList().run()"
				:class="{ 'is-active': editor.isActive('orderedList') }"
				v-tooltip="$t('input.editor.orderedList')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-list-ul']"/>
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleTaskList().run()"
				:class="{ 'is-active': editor.isActive('taskList') }"
				v-tooltip="$t('input.editor.taskList')"
			>
				<span class="icon">
					<icon icon="fa-list-check"/>
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button" 
				@click="openImagePicker" 
				v-tooltip="$t('input.editor.image')"
			>
				<span class="icon">
					<icon icon="fa-image"/>
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="setLink"
				:class="{ 'is-active': editor.isActive('link') }"
				title="set link"
				v-tooltip="$t('input.editor.link')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-link']"/>
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().setParagraph().run()"
				:class="{ 'is-active': editor.isActive('paragraph') }"
				title="paragraph"
				v-tooltip="$t('input.editor.text')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-paragraph']"/>
				</span>
			</BaseButton>

			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().setHorizontalRule().run()"
				v-tooltip="$t('input.editor.horizontalRule')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-ruler-horizontal']"/>
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().undo().run()"
				v-tooltip="$t('input.editor.undo')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-undo']"/>
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().redo().run()"
				v-tooltip="$t('input.editor.redo')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-redo']"/>
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<!-- table -->
			<BaseButton
				class="editor-toolbar__button"
				@click="toggleTableMode"
				:class="{ 'is-active': editor.isActive('table') }"
				v-tooltip="$t('input.editor.table.title')"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-table']"/>
				</span>
			</BaseButton>
			<div v-if="tableMode" class="editor-toolbar__table-buttons">
				<BaseButton
					class="editor-toolbar__button"
					@click="
						editor
							.chain()
							.focus()
							.insertTable({ rows: 3, cols: 3, withHeaderRow: true })
							.run()
					"
				>
					{{ $t('input.editor.table.insert') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().addColumnBefore().run()"
					:disabled="!editor.can().addColumnBefore"
				>
					{{ $t('input.editor.table.addColumnBefore') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().addColumnAfter().run()"
					:disabled="!editor.can().addColumnAfter"
				>
					{{ $t('input.editor.table.addColumnAfter') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().deleteColumn().run()"
					:disabled="!editor.can().deleteColumn"
				>
					{{ $t('input.editor.table.deleteColumn') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().addRowBefore().run()"
					:disabled="!editor.can().addRowBefore"
				>
					{{ $t('input.editor.table.addRowBefore') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().addRowAfter().run()"
					:disabled="!editor.can().addRowAfter"
				>
					{{ $t('input.editor.table.addRowAfter') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().deleteRow().run()"
					:disabled="!editor.can().deleteRow"
				>
					{{ $t('input.editor.table.deleteRow') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().deleteTable().run()"
					:disabled="!editor.can().deleteTable"
				>
					{{ $t('input.editor.table.deleteTable') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().mergeCells().run()"
					:disabled="!editor.can().mergeCells"
				>
					{{ $t('input.editor.table.mergeCells') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().splitCell().run()"
					:disabled="!editor.can().splitCell"
				>
					{{ $t('input.editor.table.splitCell') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().toggleHeaderColumn().run()"
					:disabled="!editor.can().toggleHeaderColumn"
				>
					{{ $t('input.editor.table.toggleHeaderColumn') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().toggleHeaderRow().run()"
					:disabled="!editor.can().toggleHeaderRow"
				>
					{{ $t('input.editor.table.toggleHeaderRow') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().toggleHeaderCell().run()"
					:disabled="!editor.can().toggleHeaderCell"
				>
					{{ $t('input.editor.table.toggleHeaderCell') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().mergeOrSplit().run()"
					:disabled="!editor.can().mergeOrSplit"
				>
					{{ $t('input.editor.table.mergeOrSplit') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().fixTables().run()"
					:disabled="!editor.can().fixTables"
				>
					{{ $t('input.editor.table.fixTables') }}
				</BaseButton>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {Editor} from '@tiptap/vue-3'

import BaseButton from '@/components/base/BaseButton.vue'

const {
	editor = null,
} = defineProps<{
	editor: Editor,
}>()

const tableMode = ref(false)

function toggleTableMode() {
	tableMode.value = !tableMode.value
}

function openImagePicker() {
	document.getElementById('tiptap__image-upload').click()
}

function setLink() {
	const previousUrl = editor.getAttributes('link').href
	const url = window.prompt('URL', previousUrl)

	// cancelled
	if (url === null) {
		return
	}

	// empty
	if (url === '') {
		editor.chain().focus().extendMarkRange('link').unsetLink().run()

		return
	}

	// update link
	editor
		.chain()
		.focus()
		.extendMarkRange('link')
		.setLink({href: url, target: '_blank'})
		.run()
}
</script>

<style lang="scss" scoped>
.editor-toolbar {
	background: var(--white);
	border: 1px solid var(--grey-200);
	user-select: none;
	padding: .5rem;
	border-radius: $radius;
	display: flex;
	flex-wrap: wrap;

	> * + * {
		border-left: 1px solid var(--grey-200);
		margin-left: 6px;
		padding-left: 6px;
	}
}

.editor-toolbar__button {
	min-width: 2rem;
	height: 2rem;
	border-radius: $radius;
	border: 1px solid transparent;
	color: var(--grey-700);
	transition: all $transition;
	background: transparent;
	margin-right: .25rem;

	&:hover {
		background: var(--grey-100);
		border-color: var(--grey-200);
	}
	
	.icon {
		position: relative;
		
		.icon__lower-text {
			font-size: .75rem;
			position: absolute;
			bottom: -3px;
			right: -2px;
			font-weight: bold;
		}
	}
}

.editor-toolbar__table-buttons {
	margin-top: .5rem;
	
	> .editor-toolbar__button {
		margin-right: .5rem;
		margin-bottom: .5rem;
		padding: 0 .25rem;
		border: 1px solid var(--grey-400);
		font-size: .75rem;
		height: 1.5rem;
	}
}
</style>
