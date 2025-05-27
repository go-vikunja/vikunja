<script setup lang="ts">
import {onBeforeUnmount, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import DatepickerWithValues from '@/components/date/DatepickerWithValues.vue'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import {
	transformFilterStringForApi,
	transformFilterStringFromApi,
} from '@/helpers/filters'
import {useDebounceFn} from '@vueuse/core'

// TipTap imports
import {EditorContent, useEditor} from '@tiptap/vue-3'
import {Extension} from '@tiptap/core'
import StarterKit from '@tiptap/starter-kit'
import {Placeholder} from '@tiptap/extension-placeholder'
import {Plugin, PluginKey} from '@tiptap/pm/state'
import {filterHighlighter} from '@/components/input/filter/highlighter.ts'
import FilterAutocomplete from '@/components/input/filter/FilterAutocomplete'

const props = defineProps<{
	projectId?: number,
	modelValue?: string,
}>()

const emit = defineEmits(['update:modelValue'])
const {t} = useI18n()

// Services and stores for autocomplete
const labelStore = useLabelStore()
const projectStore = useProjectStore()

// Date picker functionality
const currentOldDatepickerValue = ref('')
const currentDatepickerValue = ref('')
const currentDatepickerPos = ref(0)
const datePickerPopupOpen = ref(false)

// Create a custom extension for filter syntax highlighting
const FilterHighlighter = Extension.create({
	name: 'filterHighlighter',

	addProseMirrorPlugins() {
		return [
			filterHighlighter,
		]
	},
})

// Create a custom extension for handling date clicks
const DateClickHandler = Extension.create({
	name: 'dateClickHandler',

	addProseMirrorPlugins() {
		return [
			new Plugin({
				key: new PluginKey('dateClickHandler'),
				props: {
					handleClick: (view, pos, event) => {
						const target = event.target as HTMLElement
						if (target.classList.contains('date-value')) {
							event.preventDefault()
							event.stopPropagation()

							const dateValue = target.getAttribute('data-date-value') || ''
							const position = parseInt(target.getAttribute('data-position') || '0')

							currentOldDatepickerValue.value = dateValue
							currentDatepickerValue.value = dateValue
							currentDatepickerPos.value = position
							datePickerPopupOpen.value = true

							return true
						}
						return false
					},
				},
			}),
		]
	},
})

// Initialize TipTap editor
const editor = useEditor({
	extensions: [
		StarterKit.configure({
			history: false, // We'll handle history ourselves
		}),
		Placeholder.configure({
			placeholder: t('filters.query.placeholder'),
		}),
		FilterHighlighter,
		DateClickHandler,
		FilterAutocomplete.configure({
			projectId: props.projectId,
		}),
		Extension.create({
			name: 'enterHandler',
			addKeyboardShortcuts() {
				return {
					'Enter': () => {
						blurDebounced()
						return true
					},
				}
			},
		}),
	],
	content: '',
	onUpdate: ({editor}) => {
		const content = editor.getText()
		emit('update:modelValue', processContent(content))
	},
})

// Process the editor content to output snake_cased filter
const processContent = (content: string) => {
	return transformFilterStringForApi(
		content,
		labelTitle => labelStore.getLabelByExactTitle(labelTitle)?.id || null,
		projectTitle => {
			const found = projectStore.findProjectByExactname(projectTitle)
			return found?.id || null
		},
	)
}

// Watch for changes to the model value
watch(() => props.modelValue, (newValue) => {
	if (!editor.value || !newValue) return

	const content = transformFilterStringFromApi(
		newValue,
		labelId => labelStore.getLabelById(labelId)?.title || null,
		projectId => projectStore.projects[projectId]?.title || null,
	)

	// Only update if the content is different
	if (editor.value.getText() !== content) {
		editor.value.commands.setContent(content, false)
	}
}, {immediate: true})

function updateDateInQuery(newDate: string | Date | null) {
	if (!editor.value || !newDate) return

	const dateStr = typeof newDate === 'string' ? newDate : newDate.toISOString().split('T')[0]
	const currentText = editor.value.getText()
	const newText = currentText.replace(currentOldDatepickerValue.value, dateStr)
	currentOldDatepickerValue.value = dateStr

	// Update the editor content
	editor.value.commands.setContent(newText, false)
	emit('update:modelValue', processContent(newText))
}


// The blur from the editor might happen before the replacement after autocomplete select was done.
const blurDebounced = useDebounceFn(() => {
}, 500)


onBeforeUnmount(() => {
	editor.value?.destroy()
})
</script>

<template>
	<div class="filter-input">
		<div
			class="editor-wrapper"
			@blur="blurDebounced"
		>
			<EditorContent
				:editor="editor"
				class="editor-content"
			/>
		</div>
		<DatepickerWithValues
			v-model="currentDatepickerValue"
			v-model:open="datePickerPopupOpen"
			class="filter-datepicker"
			:ignore-click-classes="['date-value']"
			@update:modelValue="updateDateInQuery"
		/>
	</div>
</template>

<style lang="scss">
.filter-input {
	border: 1px solid var(--input-border-color);
	border-radius: var(--input-radius);
	background: var(--white);
	transition: border-color 0.2s ease;

	&:focus-within {
		border-color: var(--primary);
	}

	.filter-datepicker {
		position: absolute;
	}
}

.editor-wrapper {
	position: relative;
}

.editor-content {
	line-height: 1.5;
	padding: .5rem .75rem;
}

.ProseMirror {
	outline: none;
	white-space: pre-wrap;
	padding: 0 !important;

	.field {
		color: var(--code-literal);
	}

	.operator {
		color: var(--code-keyword);
	}

	.value {
		border-radius: $radius;
		padding: .125rem .25rem;
		background: var(--grey-100);
	}

	.label-value {
		border-radius: $radius;
		padding: .125rem .25rem;
	}

	.date-value {
		background-color: var(--primary);
		color: var(--white);
		border-radius: $radius;
		padding: 0.125em 0.25em;
		cursor: pointer;
		transition: background-color var(--transition);

		&:hover {
			background-color: var(--primary-dark);
		}
	}

	.grouping, .logical {
		color: var(--code-section);
	}

	.user-value {
		position: relative;
		padding-left: 1.5em;

		&::before {
			content: attr(data-user);
			position: absolute;
			left: 0;
			top: 50%;
			transform: translateY(-50%);
			width: 1.2em;
			height: 1.2em;
			background-color: #3b82f6;
			color: white;
			border-radius: 50%;
			font-size: 0.8em;
			display: flex;
			align-items: center;
			justify-content: center;
			text-transform: uppercase;
		}
	}

	p.is-editor-empty:first-child::before {
		color: var(--grey-500);
		content: attr(data-placeholder);
		float: left;
		height: 0;
		pointer-events: none;
	}
}
</style>
