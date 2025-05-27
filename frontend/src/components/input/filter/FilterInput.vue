<script setup lang="ts">
import {onBeforeUnmount, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import DatepickerWithValues from '@/components/date/DatepickerWithValues.vue'
import AutocompleteDropdown from '@/components/input/AutocompleteDropdown.vue'
import UserService from '@/services/user'
import ProjectUserService from '@/services/projectUsers'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import XLabel from '@/components/tasks/partials/Label.vue'
import User from '@/components/misc/User.vue'
import {
	ASSIGNEE_FIELDS,
	AUTOCOMPLETE_FIELDS,
	FILTER_OPERATORS_REGEX,
	LABEL_FIELDS,
	PROJECT_FIELDS,
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

const props = defineProps<{
	projectId?: number,
	modelValue?: string,
}>()

const emit = defineEmits(['update:modelValue'])
const {t} = useI18n()

// Services and stores for autocomplete
const userService = new UserService()
const projectUserService = new ProjectUserService()
const labelStore = useLabelStore()
const projectStore = useProjectStore()

// Autocomplete state
const autocompleteMatchPosition = ref(0)
const autocompleteMatchText = ref('')
const autocompleteResultType = ref<'labels' | 'assignees' | 'projects' | null>(null)
const autocompleteResults = ref<any[]>([])

// Store references to the dropdown functions
const dropdownOnFocusField = ref<(() => void) | null>(null)
const dropdownOnKeydown = ref<((event: KeyboardEvent) => void) | null>(null)

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
		handleFieldInput()
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

// Autocomplete functionality
function handleFieldInput() {
	if (!editor.value) return

	const cursorPosition = editor.value.state.selection.from
	const text = editor.value.getText()
	const textUpToCursor = text.substring(0, cursorPosition)
	autocompleteResults.value = []

	AUTOCOMPLETE_FIELDS.forEach(field => {
		const pattern = new RegExp('(' + field + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*)([\'"]?)([^\'"&|()]+\\1?)?$', 'ig')
		const match = pattern.exec(textUpToCursor)

		if (match === null) {
			return
		}

		const [matched, prefix, operator, , keyword] = match
		if (!keyword) {
			return
		}

		let search = keyword
		if (operator === 'in' || operator === '?=') {
			const keywords = keyword.split(',')
			search = keywords[keywords.length - 1].trim()
		}

		if (LABEL_FIELDS.includes(field)) {
			autocompleteResultType.value = 'labels'
			autocompleteResults.value = labelStore.filterLabelsByQuery([], search)
		}
		if (ASSIGNEE_FIELDS.includes(field)) {
			autocompleteResultType.value = 'assignees'
			if (props.projectId) {
				projectUserService.getAll({projectId: props.projectId} as any, {s: search})
					.then(users => autocompleteResults.value = users.length > 1 ? users : [])
			} else {
				userService.getAll({} as any, {s: search})
					.then(users => autocompleteResults.value = users.length > 1 ? users : [])
			}
		}
		if (!props.projectId && PROJECT_FIELDS.includes(field)) {
			autocompleteResultType.value = 'projects'
			autocompleteResults.value = projectStore.searchProject(search)
		}
		autocompleteMatchText.value = keyword
		autocompleteMatchPosition.value = match.index + prefix.length - 1 + keyword.replace(search, '').length
	})
}

function autocompleteSelect(value: any) {
	if (!editor.value) return

	const newValue = autocompleteResultType.value === 'assignees' ? value.username : value.title
	const currentText = editor.value.getText()
	const newText = currentText.substring(0, autocompleteMatchPosition.value + 1) +
		newValue +
		currentText.substring(autocompleteMatchPosition.value + autocompleteMatchText.value.length + 1)

	// Update the editor content
	editor.value.commands.setContent(newText, false)
	emit('update:modelValue', processContent(newText))

	autocompleteResults.value = []
}

// The blur from the editor might happen before the replacement after autocomplete select was done.
const blurDebounced = useDebounceFn(() => {
}, 500)

// Function to setup dropdown callbacks from the slot props
function setupDropdownCallbacks(onFocusField: () => void, onKeydown: (event: KeyboardEvent) => void) {
	dropdownOnFocusField.value = onFocusField
	dropdownOnKeydown.value = onKeydown
	return ''
}

onBeforeUnmount(() => {
	editor.value?.destroy()
})
</script>

<template>
	<div class="filter-input">
		<AutocompleteDropdown
			:options="autocompleteResults"
			@blur="editor?.commands.blur()"
			@update:modelValue="autocompleteSelect"
		>
			<template #input="{ onKeydown, onFocusField }">
				<div
					class="editor-wrapper"
					:class="{'has-autocomplete-results': autocompleteResults.length > 0}"
					@keydown="onKeydown"
					@focus="onFocusField"
					@blur="blurDebounced"
				>
					<EditorContent
						:editor="editor"
						class="editor-content"
					/>
				</div>
				<span v-show="false">{{ setupDropdownCallbacks(onFocusField, onKeydown) }}</span>
			</template>
			<template #result="{ item }">
				<XLabel
					v-if="autocompleteResultType === 'labels'"
					:label="item"
				/>
				<User
					v-else-if="autocompleteResultType === 'assignees'"
					:user="item"
					:avatar-size="25"
				/>
				<template v-else>
					{{ item.title }}
				</template>
			</template>
		</AutocompleteDropdown>
		<DatepickerWithValues
			class="filter-datepicker"
			v-model="currentDatepickerValue"
			v-model:open="datePickerPopupOpen"
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
		font-weight: 500;
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
