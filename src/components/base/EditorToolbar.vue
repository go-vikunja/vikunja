<template>
	<div class="editor-toolbar">
		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleHeading({ level: 1 }).run()"
				:class="{ 'is-active': editor.isActive('heading', { level: 1 }) }"
				title="h1"
			>
				<span class="icon"> <icon :icon="['fa', 'fa-header']" /> </span>1
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
				:class="{ 'is-active': editor.isActive('heading', { level: 2 }) }"
				title="h2"
			>
				<span class="icon"> <icon :icon="['fa', 'fa-header']" /> </span>2
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleHeading({ level: 3 }).run()"
				:class="{ 'is-active': editor.isActive('heading', { level: 3 }) }"
				title="h3"
			>
				<span class="icon"> <icon :icon="['fa', 'fa-header']" /> </span>3
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleHeading({ level: 4 }).run()"
				:class="{ 'is-active': editor.isActive('heading', { level: 4 }) }"
				title="h4"
			>
				<span class="icon"> <icon :icon="['fa', 'fa-header']" /> </span>4
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleHeading({ level: 5 }).run()"
				:class="{ 'is-active': editor.isActive('heading', { level: 5 }) }"
				title="h5"
			>
				<span class="icon"> <icon :icon="['fa', 'fa-header']" /> </span>5
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleHeading({ level: 6 }).run()"
				:class="{ 'is-active': editor.isActive('heading', { level: 6 }) }"
				title="h6"
			>
				<span class="icon"> <icon :icon="['fa', 'fa-header']" /> </span>6
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleBold().run()"
				:class="{ 'is-active': editor.isActive('bold') }"
				title="bold"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-bold']" />
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleItalic().run()"
				:class="{ 'is-active': editor.isActive('italic') }"
				title="italic"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-italic']" />
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleStrike().run()"
				:class="{ 'is-active': editor.isActive('strike') }"
				title="strike"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-strikethrough']" />
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleCode().run()"
				:class="{ 'is-active': editor.isActive('code') }"
				title="code"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-code']" />
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleCodeBlock().run()"
				:class="{ 'is-active': editor.isActive('codeBlock') }"
				title="code block"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-code']" />
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleBlockquote().run()"
				:class="{ 'is-active': editor.isActive('blockquote') }"
				title="quote"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-quote-right']" />
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleBulletList().run()"
				:class="{ 'is-active': editor.isActive('bulletList') }"
				title="bullet list"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-list-ol']" />
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleOrderedList().run()"
				:class="{ 'is-active': editor.isActive('orderedList') }"
				title="ordered list"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-list-ul']" />
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().unsetAllMarks().run()"
			>
				clear marks
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().clearNodes().run()"
			>
				clear nodes
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="setLink"
				:class="{ 'is-active': editor.isActive('link') }"
				title="set link"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-link']" />
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().unsetLink().run()"
				:disabled="!editor.isActive('link')"
				title="unset link"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-unlink']" />
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().toggleTaskList().run()"
				:class="{ 'is-active': editor.isActive('taskList') }"
				title="task list"
			>
				<span class="icon">
					<icon :icon="['fa', ' fa-list-check']" />
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().setParagraph().run()"
				:class="{ 'is-active': editor.isActive('paragraph') }"
				title="paragraph"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-paragraph']" />
				</span>
			</BaseButton>

			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().setHorizontalRule().run()"
			>
				<span class="editor-toolbar__pseudo-icon">
					-
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().setHardBreak().run()"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-arrow-turn-down']" />
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().undo().run()"
				title="undo"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-undo']" />
				</span>
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				@click="editor.chain().focus().redo().run()"
				title="redo"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-redo']" />
				</span>
			</BaseButton>
		</div>

		<div class="editor-toolbar__segment">
			<!-- table -->
			<BaseButton
				class="editor-toolbar__button"
				@click="toggleTableMode"
				:class="{ 'is-active': editor.isActive('table') }"
				title="table"
			>
				<span class="icon">
					<icon :icon="['fa', 'fa-table']" />
				</span>
			</BaseButton>
			<div v-if="tableMode">
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
					insertTable
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().addColumnBefore().run()"
					:disabled="!editor.can().addColumnBefore"
				>
					addColumnBefore
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().addColumnAfter().run()"
					:disabled="!editor.can().addColumnAfter"
				>
					addColumnAfter
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().deleteColumn().run()"
					:disabled="!editor.can().deleteColumn"
				>
					deleteColumn
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().addRowBefore().run()"
					:disabled="!editor.can().addRowBefore"
				>
					addRowBefore
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().addRowAfter().run()"
					:disabled="!editor.can().addRowAfter"
				>
					addRowAfter
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().deleteRow().run()"
					:disabled="!editor.can().deleteRow"
				>
					deleteRow
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().deleteTable().run()"
					:disabled="!editor.can().deleteTable"
				>
					deleteTable
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().mergeCells().run()"
					:disabled="!editor.can().mergeCells"
				>
					mergeCells
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().splitCell().run()"
					:disabled="!editor.can().splitCell"
				>
					splitCell
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().toggleHeaderColumn().run()"
					:disabled="!editor.can().toggleHeaderColumn"
				>
					toggleHeaderColumn
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().toggleHeaderRow().run()"
					:disabled="!editor.can().toggleHeaderRow"
				>
					toggleHeaderRow
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().toggleHeaderCell().run()"
					:disabled="!editor.can().toggleHeaderCell"
				>
					toggleHeaderCell
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().mergeOrSplit().run()"
					:disabled="!editor.can().mergeOrSplit"
				>
					mergeOrSplit
				</BaseButton>
				<!-- <BaseButton
					class="editor-toolbar__button"
					@click="
						editor
							.chain()
							.focus()
							.setCellAttribute('backgroundColor', '#FAF594')
							.run()
					"
					:disabled="
						!editor.can().setCellAttribute('backgroundColor', '#FAF594')
					"
				>
					setCellAttribute
				</BaseButton> -->
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().fixTables().run()"
					:disabled="!editor.can().fixTables"
				>
					fixTables
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().goToNextCell().run()"
					:disabled="!editor.can().goToNextCell"
				>
					goToNextCell
				</BaseButton>
				<BaseButton
					class="editor-toolbar__button"
					@click="editor.chain().focus().goToPreviousCell().run()"
					:disabled="!editor.can().goToPreviousCell"
				>
					goToPreviousCell
				</BaseButton>
			</div>
		</div>

		<div class="editor-toolbar__segment">
			<BaseButton class="editor-toolbar__button" @click="addImage" title="Add image from URL">
				<span class="icon">
					<icon icon="fa-image" />
				</span>
			</BaseButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref, type PropType} from 'vue'
import {Editor} from '@tiptap/vue-3'

import BaseButton from '@/components/base/BaseButton.vue'

const props = defineProps({
	editor: {
		default: null,
		type: Editor as PropType<Editor>,
	},
})

const tableMode = ref(false)

function toggleTableMode() {
	tableMode.value = !tableMode.value
}

function addImage() {
	const url = window.prompt('URL')

	if (url) {
		props.editor?.chain().focus().setImage({ src: url }).run()
	}
}

function setLink() {
	const previousUrl = props.editor.getAttributes('link').href
	const url = window.prompt('URL', previousUrl)

	// cancelled
	if (url === null) {
		return
	}

	// empty
	if (url === '') {
		props.editor.chain().focus().extendMarkRange('link').unsetLink().run()

		return
	}

	// update link
	props.editor
		.chain()
		.focus()
		.extendMarkRange('link')
		.setLink({ href: url, target: '_blank' })
		.run()
}
</script>

<style lang="scss" scoped>
.editor-toolbar {
	background: var(--grey-50);
	border: 1px solid var(--grey-200);
	border-bottom: none;
	// position: relative;
	user-select: none;
	padding: 9px 10px;
	border-top-left-radius: 4px;
	border-top-right-radius: 4px;
	display: flex;
	flex-wrap: wrap;

	> * + * {
		// .editor-toolbar i.separator {
		border--left-color: var(--grey-200) !important;
		// }
		// .editor-toolbar i.separator {
		// display: inline-block;
		// width: 0;
		border-left: 1px solid var(--grey-200);
		// border-right: 1px solid #fff;
		// color: transparent;
		// text-indent: -10px;
		margin-left: 6px;
		padding-left: 6px;
		// }
	}
}

.editor-toolbar__button {
	color: var(--grey-700);
	// width: 30px;
	height: 30px;
	border-radius: 3px;
	border: 1px solid transparent;

	&:hover {
		background: var(--grey-200);
		border-color: var(--grey-300);
	}
}

.editor-toolbar__pseudo-icon {
	padding: 0 .5rem;
	font-weight: bold;
	vertical-align: top;
}
</style>
