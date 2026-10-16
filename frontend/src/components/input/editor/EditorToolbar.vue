<template>
	<div
		ref="toolbarRef"
		class="editor-toolbar"
		role="toolbar"
		:aria-label="$t('input.editor.toolbarLabel')"
		@keydown="onToolbarKeydown"
		@focusin="onToolbarFocusin"
	>
		<div
			ref="rowRef"
			class="editor-toolbar__row"
		>
			<BaseButton
				v-for="(action, index) in actions"
				:key="action.label"
				v-tooltip="$t(action.label)"
				class="editor-toolbar__button"
				:class="{
					'is-active': action.isActive?.(),
					'is-group-start': action.startsGroup && index > 0,
					'is-overflowing': index >= firstOverflowIndex,
				}"
				:aria-pressed="action.isActive?.()"
				@click="action.run"
			>
				<span class="icon">
					<Icon :icon="action.icon" />
					<span
						v-if="action.headingLevel"
						class="icon__lower-text"
						aria-hidden="true"
					>{{ action.headingLevel }}</span>
				</span>
				<span class="is-sr-only">{{ $t(action.label) }}</span>
			</BaseButton>
		</div>

		<Dropdown
			class="editor-toolbar__more"
			:class="{'is-hidden-more': overflowingActions.length === 0}"
		>
			<template #trigger="{toggleOpen, open}">
				<BaseButton
					v-tooltip="$t('input.editor.more')"
					class="editor-toolbar__button"
					aria-haspopup="true"
					:aria-expanded="open"
					@click="(event: MouseEvent) => onMoreClick(event, toggleOpen, open)"
					@keydown.down.prevent="() => onMoreClick(undefined, toggleOpen, open)"
				>
					<span class="icon">
						<Icon icon="ellipsis-h" />
					</span>
					<span class="is-sr-only">{{ $t('input.editor.more') }}</span>
				</BaseButton>
			</template>
			<template #default="{close}">
				<template
					v-for="(action, index) in overflowingActions"
					:key="action.label"
				>
					<hr
						v-if="action.startsGroup && index > 0"
						class="dropdown-divider"
					>
					<DropdownItem
						:icon="action.icon"
						:aria-pressed="action.isActive?.()"
						@click="(event: MouseEvent) => { action.run(event); close() }"
					>
						{{ $t(action.label) }}
						<Icon
							v-if="action.isActive?.()"
							icon="check"
							class="editor-toolbar__menu-check"
						/>
					</DropdownItem>
				</template>
			</template>
		</Dropdown>

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
				:disabled="!editor.can().addColumnBefore()"
				@click="editor.chain().focus().addColumnBefore().run()"
			>
				{{ $t('input.editor.table.addColumnBefore') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().addColumnAfter()"
				@click="editor.chain().focus().addColumnAfter().run()"
			>
				{{ $t('input.editor.table.addColumnAfter') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().deleteColumn()"
				@click="editor.chain().focus().deleteColumn().run()"
			>
				{{ $t('input.editor.table.deleteColumn') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().addRowBefore()"
				@click="editor.chain().focus().addRowBefore().run()"
			>
				{{ $t('input.editor.table.addRowBefore') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().addRowAfter()"
				@click="editor.chain().focus().addRowAfter().run()"
			>
				{{ $t('input.editor.table.addRowAfter') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().deleteRow()"
				@click="editor.chain().focus().deleteRow().run()"
			>
				{{ $t('input.editor.table.deleteRow') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().deleteTable()"
				@click="editor.chain().focus().deleteTable().run()"
			>
				{{ $t('input.editor.table.deleteTable') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().mergeCells()"
				@click="editor.chain().focus().mergeCells().run()"
			>
				{{ $t('input.editor.table.mergeCells') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().splitCell()"
				@click="editor.chain().focus().splitCell().run()"
			>
				{{ $t('input.editor.table.splitCell') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().toggleHeaderColumn()"
				@click="editor.chain().focus().toggleHeaderColumn().run()"
			>
				{{ $t('input.editor.table.toggleHeaderColumn') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().toggleHeaderRow()"
				@click="editor.chain().focus().toggleHeaderRow().run()"
			>
				{{ $t('input.editor.table.toggleHeaderRow') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().toggleHeaderCell()"
				@click="editor.chain().focus().toggleHeaderCell().run()"
			>
				{{ $t('input.editor.table.toggleHeaderCell') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().mergeOrSplit()"
				@click="editor.chain().focus().mergeOrSplit().run()"
			>
				{{ $t('input.editor.table.mergeOrSplit') }}
			</BaseButton>
			<BaseButton
				class="editor-toolbar__button"
				:disabled="!editor.can().fixTables()"
				@click="editor.chain().focus().fixTables().run()"
			>
				{{ $t('input.editor.table.fixTables') }}
			</BaseButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch, nextTick, onMounted, onBeforeUnmount} from 'vue'
import {useResizeObserver} from '@vueuse/core'
import type {Editor} from '@tiptap/vue-3'
import type {IconProp} from '@fortawesome/fontawesome-svg-core'

import BaseButton from '@/components/base/BaseButton.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import {setLinkInEditor} from '@/components/input/editor/setLinkInEditor'

const props = defineProps<{
	editor: Editor,
}>()

const emit = defineEmits(['imageUploadClicked'])

const tableMode = ref(false)

function toggleTableMode() {
	tableMode.value = !tableMode.value
}

interface ToolbarAction {
	label: string
	icon: IconProp
	headingLevel?: 1 | 2 | 3
	startsGroup?: boolean
	isActive?: () => boolean
	run: (event: MouseEvent) => void
}

const actions: ToolbarAction[] = [
	...([1, 2, 3] as const).map((level): ToolbarAction => ({
		label: `input.editor.heading${level}`,
		icon: ['fas', 'header'],
		headingLevel: level,
		startsGroup: level === 1,
		isActive: () => props.editor.isActive('heading', {level}),
		run: () => props.editor.chain().focus().toggleHeading({level}).run(),
	})),
	{
		label: 'input.editor.bold',
		icon: ['fas', 'bold'],
		startsGroup: true,
		isActive: () => props.editor.isActive('bold'),
		run: () => props.editor.chain().focus().toggleBold().run(),
	},
	{
		label: 'input.editor.italic',
		icon: ['fas', 'italic'],
		isActive: () => props.editor.isActive('italic'),
		run: () => props.editor.chain().focus().toggleItalic().run(),
	},
	{
		label: 'input.editor.underline',
		icon: ['fas', 'underline'],
		isActive: () => props.editor.isActive('underline'),
		run: () => props.editor.chain().focus().toggleUnderline().run(),
	},
	{
		label: 'input.editor.strikethrough',
		icon: ['fas', 'strikethrough'],
		isActive: () => props.editor.isActive('strike'),
		run: () => props.editor.chain().focus().toggleStrike().run(),
	},
	{
		label: 'input.editor.code',
		icon: ['fas', 'code'],
		startsGroup: true,
		isActive: () => props.editor.isActive('codeBlock'),
		run: () => props.editor.chain().focus().toggleCodeBlock().run(),
	},
	{
		label: 'input.editor.quote',
		icon: ['fas', 'quote-right'],
		isActive: () => props.editor.isActive('blockquote'),
		run: () => props.editor.chain().focus().toggleBlockquote().run(),
	},
	{
		label: 'input.editor.bulletList',
		icon: ['fas', 'list-ul'],
		startsGroup: true,
		isActive: () => props.editor.isActive('bulletList'),
		run: () => props.editor.chain().focus().toggleBulletList().run(),
	},
	{
		label: 'input.editor.orderedList',
		icon: ['fas', 'list-ol'],
		isActive: () => props.editor.isActive('orderedList'),
		run: () => props.editor.chain().focus().toggleOrderedList().run(),
	},
	{
		label: 'input.editor.taskList',
		icon: 'list-check',
		isActive: () => props.editor.isActive('taskList'),
		run: () => props.editor.chain().focus().toggleTaskList().run(),
	},
	{
		label: 'input.editor.image',
		icon: 'image',
		startsGroup: true,
		run: event => emit('imageUploadClicked', event),
	},
	{
		label: 'input.editor.link',
		icon: ['fas', 'link'],
		startsGroup: true,
		isActive: () => props.editor.isActive('link'),
		run: event => setLinkInEditor((event.currentTarget as HTMLElement).getBoundingClientRect(), props.editor),
	},
	{
		label: 'input.editor.text',
		icon: ['fas', 'paragraph'],
		isActive: () => props.editor.isActive('paragraph'),
		run: () => props.editor.chain().focus().setParagraph().run(),
	},
	{
		label: 'input.editor.horizontalRule',
		icon: ['fas', 'ruler-horizontal'],
		run: () => props.editor.chain().focus().setHorizontalRule().run(),
	},
	{
		label: 'input.editor.undo',
		icon: ['fas', 'undo'],
		startsGroup: true,
		run: () => props.editor.chain().focus().undo().run(),
	},
	{
		label: 'input.editor.redo',
		icon: ['fas', 'redo'],
		run: () => props.editor.chain().focus().redo().run(),
	},
	{
		label: 'input.editor.table.title',
		icon: ['fas', 'table'],
		startsGroup: true,
		isActive: () => props.editor.isActive('table'),
		run: toggleTableMode,
	},
]

// Overflowing buttons wrap onto clipped rows but keep their slot, so hiding them
// can't change how many fit.
const rowRef = ref<HTMLElement | null>(null)
const firstOverflowIndex = ref(actions.length)
const overflowingActions = computed(() => actions.slice(firstOverflowIndex.value))

function updateOverflow() {
	const buttons = Array.from(rowRef.value?.children ?? []) as HTMLElement[]
	const firstRowTop = buttons[0]?.offsetTop
	const index = buttons.findIndex(button => button.offsetTop > firstRowTop)
	firstOverflowIndex.value = index === -1 ? actions.length : index
}

useResizeObserver(rowRef, updateOverflow)

watch(firstOverflowIndex, async () => {
	await nextTick()
	setRovingTabindex(document.activeElement as HTMLElement)
})

function menuItems(): HTMLButtonElement[] {
	return Array.from(toolbarRef.value?.querySelectorAll<HTMLButtonElement>('.dropdown-menu button:not(:disabled)') ?? [])
}

async function onMoreClick(event: MouseEvent | undefined, toggleOpen: () => void, open: boolean) {
	if (event === undefined && open) {
		menuItems()[0]?.focus()
		return
	}
	toggleOpen()
	// detail is 0 for keyboard clicks
	if (!open && (event === undefined || event.detail === 0)) {
		await nextTick()
		menuItems()[0]?.focus()
	}
}

// Roving tabindex: the toolbar is a single Tab stop and arrow keys move between
// buttons, so Tab always leaves the toolbar (to the editor content) instead of
// walking through every button.
const toolbarRef = ref<HTMLElement | null>(null)

function allToolbarButtons(): HTMLButtonElement[] {
	if (!toolbarRef.value) {
		return []
	}
	return Array.from(toolbarRef.value.querySelectorAll<HTMLButtonElement>('button'))
		.filter(button => !button.closest('.dropdown-menu'))
}

function toolbarButtons(): HTMLButtonElement[] {
	return allToolbarButtons()
		.filter(button => !button.disabled && getComputedStyle(button).visibility !== 'hidden')
}

function setRovingTabindex(active: HTMLElement | null) {
	const buttons = toolbarButtons()
	if (buttons.length === 0) {
		return
	}
	const target = active && buttons.includes(active as HTMLButtonElement) ? active : buttons[0]
	allToolbarButtons().forEach(button => {
		button.tabIndex = button === target ? 0 : -1
	})
}

function onToolbarFocusin(event: FocusEvent) {
	const target = event.target as HTMLElement
	if (target.tagName === 'BUTTON' && !target.closest('.dropdown-menu')) {
		setRovingTabindex(target)
	}
}

function nextIndex(key: string, index: number, length: number, forward: string, backward: string): number | null {
	switch (key) {
		case forward:
			return (index + 1) % length
		case backward:
			return (index - 1 + length) % length
		case 'Home':
			return 0
		case 'End':
			return length - 1
	}
	return null
}

function onToolbarKeydown(event: KeyboardEvent) {
	const inMenu = (event.target as HTMLElement).closest('.dropdown-menu') !== null
	const buttons = inMenu ? menuItems() : toolbarButtons()
	if (buttons.length === 0) {
		return
	}

	const currentIndex = buttons.indexOf(document.activeElement as HTMLButtonElement)
	const next = nextIndex(
		event.key,
		currentIndex === -1 ? 0 : currentIndex,
		buttons.length,
		inMenu ? 'ArrowDown' : 'ArrowRight',
		inMenu ? 'ArrowUp' : 'ArrowLeft',
	)
	if (next === null) {
		return
	}

	event.preventDefault()

	const nextButton = buttons[next]
	if (!inMenu) {
		setRovingTabindex(nextButton)
	}
	nextButton.focus()
}

// Buttons appear/disappear (table mode) and disable as the selection moves; re-assert the
// roving tabindex so Tab can always re-enter. Filtered to 'disabled' — we write tabindex ourselves.
let observer: MutationObserver | null = null

onMounted(() => {
	setRovingTabindex(null)
	observer = new MutationObserver(() => setRovingTabindex(document.activeElement as HTMLElement))
	if (toolbarRef.value) {
		observer.observe(toolbarRef.value, {
			subtree: true,
			childList: true,
			attributeFilter: ['disabled'],
		})
	}
})

onBeforeUnmount(() => {
	observer?.disconnect()
	observer = null
})
</script>

<style lang="scss" scoped>
// Keeps the focus ring inside the clipped row
$focus-ring-room: 2px;

.editor-toolbar {
	background: var(--white);
	border: 1px solid var(--grey-200);
	user-select: none;
	padding: .5rem;
	border-radius: $radius;
	display: flex;
	flex-wrap: wrap;
	align-items: flex-start;
}

.editor-toolbar__row {
	flex: 1 1 0;
	min-inline-size: 0;
	display: flex;
	flex-wrap: wrap;
	align-content: flex-start;
	position: relative;
	overflow: hidden;
	block-size: calc(2rem + #{$focus-ring-room * 2});
	padding: $focus-ring-room;
	margin: -$focus-ring-room;
}

.editor-toolbar__more.is-hidden-more {
	visibility: hidden;
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
	position: relative;

	&:hover {
		background: var(--grey-100);
		border-color: var(--grey-200);
	}

	&.is-group-start {
		margin-inline-start: 13px;

		&::before {
			content: '';
			position: absolute;
			inset-block: -1px;
			inset-inline-start: -8px;
			border-inline-start: 1px solid var(--grey-200);
		}
	}

	&.is-overflowing {
		visibility: hidden;
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

.editor-toolbar__menu-check {
	margin-inline-start: .5rem;
	color: var(--primary);
}

.editor-toolbar__table-buttons {
	flex-basis: 100%;
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
