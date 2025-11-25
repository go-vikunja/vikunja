<template>
	<div class="editor-toolbar">
		<div class="editor-toolbar__segment">
			<BaseButton
				v-tooltip="$t('input.editor.heading1')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('heading', { level: 1 }) }"
				@click="editor.chain().focus().toggleHeading({ level: 1 }).run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-header']" />
					<span
						class="icon__lower-text"
						aria-hidden="true"
					>1</span>
				</span>
				<span class="is-sr-only">{{ $t('input.editor.heading1') }}</span>
			</BaseButton>
			<BaseButton
				v-tooltip="$t('input.editor.heading2')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('heading', { level: 2 }) }"
				@click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-header']" />
					<span
						class="icon__lower-text"
						aria-hidden="true"
					>2</span>
				</span>
				<span class="is-sr-only">{{ $t('input.editor.heading2') }}</span>
			</BaseButton>
			<BaseButton
				v-tooltip="$t('input.editor.heading3')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('heading', { level: 3 }) }"
				@click="editor.chain().focus().toggleHeading({ level: 3 }).run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-header']" />
					<span
						class="icon__lower-text"
						aria-hidden="true"
					>3</span>
				</span>
				<span class="is-sr-only">{{ $t('input.editor.heading3') }}</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				v-tooltip="$t('input.editor.bold')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('bold') }"
				@click="editor.chain().focus().toggleBold().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-bold']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.bold') }}</span>
			</BaseButton>
			<BaseButton
				v-tooltip="$t('input.editor.italic')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('italic') }"
				@click="editor.chain().focus().toggleItalic().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-italic']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.italic') }}</span>
			</BaseButton>
			<BaseButton
				v-tooltip="$t('input.editor.underline')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('underline') }"
				@click="editor.chain().focus().toggleUnderline().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-underline']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.underline') }}</span>
			</BaseButton>
			<BaseButton
				v-tooltip="$t('input.editor.strikethrough')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('strike') }"
				@click="editor.chain().focus().toggleStrike().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-strikethrough']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.strikethrough') }}</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				v-tooltip="$t('input.editor.code')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('codeBlock') }"
				@click="editor.chain().focus().toggleCodeBlock().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-code']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.code') }}</span>
			</BaseButton>
			<BaseButton
				v-tooltip="$t('input.editor.quote')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('blockquote') }"
				@click="editor.chain().focus().toggleBlockquote().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-quote-right']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.quote') }}</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				v-tooltip="$t('input.editor.bulletList')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('bulletList') }"
				@click="editor.chain().focus().toggleBulletList().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-list-ul']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.bulletList') }}</span>
			</BaseButton>
			<BaseButton
				v-tooltip="$t('input.editor.orderedList')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('orderedList') }"
				@click="editor.chain().focus().toggleOrderedList().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-list-ol']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.orderedList') }}</span>
			</BaseButton>
			<BaseButton
				v-tooltip="$t('input.editor.taskList')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('taskList') }"
				@click="editor.chain().focus().toggleTaskList().run()"
			>
				<span class="icon">
					<Icon icon="fa-list-check" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.taskList') }}</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				v-tooltip="$t('input.editor.image')"
				class="editor-toolbar__button"
				@click="e => emit('imageUploadClicked', e)"
			>
				<span class="icon">
					<Icon icon="fa-image" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.image') }}</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				v-tooltip="$t('input.editor.link')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('link') }"
				title="set link"
				@click="setLink"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-link']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.link') }}</span>
			</BaseButton>
			<BaseButton
				v-tooltip="$t('input.editor.text')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('paragraph') }"
				title="paragraph"
				@click="editor.chain().focus().setParagraph().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-paragraph']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.text') }}</span>
			</BaseButton>

			<BaseButton
				v-tooltip="$t('input.editor.horizontalRule')"
				class="editor-toolbar__button"
				@click="editor.chain().focus().setHorizontalRule().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-ruler-horizontal']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.horizontalRule') }}</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				v-tooltip="$t('input.editor.undo')"
				class="editor-toolbar__button"
				@click="editor.chain().focus().undo().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-undo']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.undo') }}</span>
			</BaseButton>
			<BaseButton
				v-tooltip="$t('input.editor.redo')"
				class="editor-toolbar__button"
				@click="editor.chain().focus().redo().run()"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-redo']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.redo') }}</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<!-- table -->
			<BaseButton
				v-tooltip="$t('input.editor.table.title')"
				class="editor-toolbar__button"
				:class="{ 'is-active': editor.isActive('table') }"
				@click="toggleTableMode"
			>
				<span class="icon">
					<Icon :icon="['fa', 'fa-table']" />
				</span>
				<span class="is-sr-only">{{ $t('input.editor.table.title') }}</span>
			</BaseButton>
			<div
				v-if="tableMode"
				class="editor-toolbar__table-buttons"
			>
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
					:disabled="!editor.can().addColumnBefore"
					@click="editor.chain().focus().addColumnBefore().run()"
				>
					{{ $t('input.editor.table.addColumnBefore') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().addColumnAfter"
					@click="editor.chain().focus().addColumnAfter().run()"
				>
					{{ $t('input.editor.table.addColumnAfter') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().deleteColumn"
					@click="editor.chain().focus().deleteColumn().run()"
				>
					{{ $t('input.editor.table.deleteColumn') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().addRowBefore"
					@click="editor.chain().focus().addRowBefore().run()"
				>
					{{ $t('input.editor.table.addRowBefore') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().addRowAfter"
					@click="editor.chain().focus().addRowAfter().run()"
				>
					{{ $t('input.editor.table.addRowAfter') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().deleteRow"
					@click="editor.chain().focus().deleteRow().run()"
				>
					{{ $t('input.editor.table.deleteRow') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().deleteTable"
					@click="editor.chain().focus().deleteTable().run()"
				>
					{{ $t('input.editor.table.deleteTable') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().mergeCells"
					@click="editor.chain().focus().mergeCells().run()"
				>
					{{ $t('input.editor.table.mergeCells') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().splitCell"
					@click="editor.chain().focus().splitCell().run()"
				>
					{{ $t('input.editor.table.splitCell') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().toggleHeaderColumn"
					@click="editor.chain().focus().toggleHeaderColumn().run()"
				>
					{{ $t('input.editor.table.toggleHeaderColumn') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().toggleHeaderRow"
					@click="editor.chain().focus().toggleHeaderRow().run()"
				>
					{{ $t('input.editor.table.toggleHeaderRow') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().toggleHeaderCell"
					@click="editor.chain().focus().toggleHeaderCell().run()"
				>
					{{ $t('input.editor.table.toggleHeaderCell') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().mergeOrSplit"
					@click="editor.chain().focus().mergeOrSplit().run()"
				>
					{{ $t('input.editor.table.mergeOrSplit') }}
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					:disabled="!editor.can().fixTables"
					@click="editor.chain().focus().fixTables().run()"
				>
					{{ $t('input.editor.table.fixTables') }}
				</BaseButton>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import type {Editor} from '@tiptap/vue-3'

import BaseButton from '@/components/base/BaseButton.vue'
import {setLinkInEditor} from '@/components/input/editor/setLinkInEditor'

const props = defineProps<{
	editor: Editor,
}>()

const emit = defineEmits(['imageUploadClicked'])

const tableMode = ref(false)

function toggleTableMode() {
	tableMode.value = !tableMode.value
}

function setLink(event) {
	setLinkInEditor(event.target.getBoundingClientRect(), props.editor)
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
		border-inline-start: 1px solid var(--grey-200);
		margin-inline-start: 6px;
		padding-inline-start: 6px;
	}
}

.editor-toolbar__button {
	min-inline-size: 2rem;
	block-size: 2rem;
	border-radius: $radius;
	border: 1px solid transparent;
	color: var(--grey-700);
	transition: all $transition;
	background: transparent;
	margin-inline-end: .25rem;

	&:hover {
		background: var(--grey-100);
		border-color: var(--grey-200);
	}

	.icon {
		position: relative;

		.icon__lower-text {
			font-size: .75rem;
			position: absolute;
			inset-block-end: -3px;
			inset-inline-end: -2px;
			font-weight: bold;
		}
	}
}

.editor-toolbar__table-buttons {
	margin-block-start: .5rem;

	> .editor-toolbar__button {
		margin-inline-end: .5rem;
		margin-block-end: .5rem;
		padding: 0 .25rem;
		border: 1px solid var(--grey-400);
		font-size: .75rem;
		block-size: 1.5rem;
	}
}
</style>
