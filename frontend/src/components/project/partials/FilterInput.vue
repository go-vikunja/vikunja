<script setup lang="ts">
import {computed, nextTick, ref, watch} from 'vue'
import {useAutoHeightTextarea} from '@/composables/useAutoHeightTextarea'
import DatepickerWithValues from '@/components/date/DatepickerWithValues.vue'
import UserService from '@/services/user'
import AutocompleteDropdown from '@/components/input/AutocompleteDropdown.vue'
import {useLabelStore} from '@/stores/labels'
import XLabel from '@/components/tasks/partials/Label.vue'
import User from '@/components/misc/User.vue'
import ProjectUserService from '@/services/projectUsers'
import {useProjectStore} from '@/stores/projects'
import {
	ASSIGNEE_FIELDS,
	AUTOCOMPLETE_FIELDS,
	AVAILABLE_FILTER_FIELDS,
	DATE_FIELDS,
	FILTER_JOIN_OPERATOR,
	FILTER_OPERATORS,
	FILTER_OPERATORS_REGEX,
	getFilterFieldRegexPattern,
	LABEL_FIELDS,
} from '@/helpers/filters'
import {useDebounceFn} from '@vueuse/core'
import {createRandomID} from '@/helpers/randomId'

const props = defineProps<{
	modelValue: string,
	projectId?: number,
	inputLabel?: string,
}>()

const emit = defineEmits<{
	'update:modelValue': [value: string],
	'blur': [],
}>()

const userService = new UserService()
const projectUserService = new ProjectUserService()
const labelStore = useLabelStore()
const projectStore = useProjectStore()

const filterQuery = ref<string>('')
const {
	textarea: filterInput,
	height,
} = useAutoHeightTextarea(filterQuery)

const id = ref(createRandomID())

watch(
	() => props.modelValue,
	() => {
		filterQuery.value = props.modelValue
	},
	{immediate: true},
)

watch(
	() => filterQuery.value,
	() => {
		if (filterQuery.value !== props.modelValue) {
			emit('update:modelValue', filterQuery.value)
		}
	},
)

function escapeHtml(unsafe: string|null|undefined): string {
	if (!unsafe) {
		return ''
	}
	return unsafe
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#039;')
}

function unEscapeHtml(unsafe: string|null|undefined): string {
	if (!unsafe) {
		return ''
	}
	return unsafe
		.replace(/&amp;/g, '&')
		.replace(/&lt;/g, '<')
		.replace(/&gt;/g, '>')
		.replace(/&quot/g, '"')
		.replace(/&#039;/g, '\'')
}

const highlightedFilterQuery = computed(() => {
	if (filterQuery.value === '') {
		return ''
	}

	let highlighted = escapeHtml(filterQuery.value)
	DATE_FIELDS
		.forEach(o => {
			const pattern = new RegExp(o + '(\\s*)' + FILTER_OPERATORS_REGEX + '(\\s*)([\'"]?)([^\'"\\s]+\\1?)?', 'ig')
			highlighted = highlighted.replaceAll(pattern, (match, spacesBefore, token, spacesAfter, start, value, position) => {
				if (typeof value === 'undefined') {
					value = ''
				}

				let endPadding = ''
				if (value.endsWith(' ')) {
					const fullLength = value.length
					value = value.trimEnd()
					const numberOfRemovedSpaces = fullLength - value.length
					endPadding = endPadding.padEnd(numberOfRemovedSpaces, ' ')
				}

				return `${o}${spacesBefore}${token}${spacesAfter}<button class="is-primary filter-query__date_value" data-position="${position}">${value}</button><span class="filter-query__date_value_placeholder">${value}</span>${endPadding}`
			})
		})
	ASSIGNEE_FIELDS
		.forEach(f => {
			const pattern = new RegExp(f + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*([\'"]?)([^\'"\\s]+\\1?)?', 'ig')
			highlighted = highlighted.replaceAll(pattern, (match, token, start, value) => {
				if (typeof value === 'undefined') {
					value = ''
				}

				return `${f} ${token} <span class="filter-query__assignee_value">${value}<span>`
			})
		})
	FILTER_JOIN_OPERATOR
		.map(o => escapeHtml(o))
		.forEach(o => {
			highlighted = highlighted.replaceAll(o, `<span class="filter-query__join-operator">${o}</span>`)
		})
	LABEL_FIELDS
		.forEach(f => {
			const pattern = getFilterFieldRegexPattern(f)
			highlighted = highlighted.replaceAll(pattern, (match, prefix, operator, space, value) => {

				if (typeof value === 'undefined') {
					value = ''
				}

				let labelTitles = [value.trim()]
				if (operator === 'in' || operator === '?=' || operator === 'not in' || operator === '?!=') {
					labelTitles = value.split(',').map(v => v.trim())
				}

				const labelsHtml: string[] = []
				labelTitles.forEach(t => {
					const label = labelStore.getLabelByExactTitle(t) || undefined
					labelsHtml.push(`<span class="filter-query__label_value" style="background-color: ${label?.hexColor}; color: ${label?.textColor}">${label?.title ?? t}</span>`)
				})

				const endSpace = value.endsWith(' ') ? ' ' : ''
				return `${f} ${operator} ${labelsHtml.join(', ')}${endSpace}`
			})
		})
	FILTER_OPERATORS
		.map(o => ` ${escapeHtml(o)} `)
		.forEach(o => {
			highlighted = highlighted.replaceAll(o, `<span class="filter-query__operator">${o}</span>`)
		})
	AVAILABLE_FILTER_FIELDS.forEach(f => {
		highlighted = highlighted.replaceAll(f, `<span class="filter-query__field">${f}</span>`)
	})
	return highlighted
})

const currentOldDatepickerValue = ref('')
const currentDatepickerValue = ref('')
const currentDatepickerPos = ref()
const datePickerPopupOpen = ref(false)

watch(
	() => highlightedFilterQuery.value,
	async () => {
		await nextTick()
		document.querySelectorAll('button.filter-query__date_value')
			.forEach(b => {
				b.addEventListener('click', event => {
					event.preventDefault()
					event.stopPropagation()

					const button = event.target
					currentOldDatepickerValue.value = button?.innerText
					currentDatepickerValue.value = button?.innerText
					currentDatepickerPos.value = parseInt(button?.dataset.position)
					datePickerPopupOpen.value = true
				})
			})
	},
	{immediate: true},
)

function updateDateInQuery(newDate: string) {
	// Need to escape and unescape the query because the positions are based on the escaped query
	let escaped = escapeHtml(filterQuery.value)
	escaped = escaped
		.substring(0, currentDatepickerPos.value)
		+ escaped
			.substring(currentDatepickerPos.value)
			.replace(currentOldDatepickerValue.value, newDate)
	currentOldDatepickerValue.value = newDate
	filterQuery.value = unEscapeHtml(escaped)
}

const autocompleteMatchPosition = ref(0)
const autocompleteMatchText = ref('')
const autocompleteResultType = ref<'labels' | 'assignees' | 'projects' | null>(null)
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const autocompleteResults = ref<any[]>([])

function handleFieldInput() {
	if (!filterInput.value) return
	const cursorPosition = filterInput.value.selectionStart
	const textUpToCursor = filterQuery.value.substring(0, cursorPosition)
	autocompleteResults.value = []

	AUTOCOMPLETE_FIELDS.forEach(field => {
		const pattern = new RegExp('(' + field + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*)([\'"]?)([^\'"&|()]+\\1?)?$', 'ig')
		const match = pattern.exec(textUpToCursor)

		if (match === null) {
			return
		}

		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		const [matched, prefix, operator, space, keyword] = match
		if(!keyword) {
			return
		}

		let search = keyword
		if (operator === 'in' || operator === '?=') {
			const keywords = keyword.split(',')
			search = keywords[keywords.length - 1].trim()
		}
		if (matched.startsWith('label')) {
			autocompleteResultType.value = 'labels'
			autocompleteResults.value = labelStore.filterLabelsByQuery([], search)
		}
		if (matched.startsWith('assignee')) {
			autocompleteResultType.value = 'assignees'
			if (props.projectId) {
				projectUserService.getAll({projectId: props.projectId}, {s: search})
					.then(users => autocompleteResults.value = users.length > 1 ? users : [])
			} else {
				userService.getAll({}, {s: search})
					.then(users => autocompleteResults.value = users.length > 1 ? users : [])
			}
		}
		if (!props.projectId && matched.startsWith('project')) {
			autocompleteResultType.value = 'projects'
			autocompleteResults.value = projectStore.searchProject(search)
		}
		autocompleteMatchText.value = keyword
		autocompleteMatchPosition.value = match.index + prefix.length - 1 + keyword.replace(search, '').length
	})
}

function autocompleteSelect(value) {
	filterQuery.value = filterQuery.value.substring(0, autocompleteMatchPosition.value + 1) +
		(autocompleteResultType.value === 'assignees'
			? value.username
			: value.title) +
		filterQuery.value.substring(autocompleteMatchPosition.value + autocompleteMatchText.value.length + 1)

	autocompleteResults.value = []
}

// The blur from the textarea might happen before the replacement after autocomplete select was done.
// That caused listeners to try and replace values earlier, resulting in broken queries.
const blurDebounced = useDebounceFn(() => emit('blur'), 500)
</script>

<!-- eslint-disable vue/no-v-html -->
<template>
	<div class="field">
		<label
			class="label"
			:for="id"
		>
			{{ inputLabel ?? $t('filters.query.title') }}
		</label>
		<AutocompleteDropdown
			:options="autocompleteResults"
			@blur="filterInput?.blur()"
			@update:modelValue="autocompleteSelect"
		>
			<template
				#input="{ onKeydown, onFocusField }"
			>
				<div class="control filter-input">
					<textarea
						:id
						ref="filterInput"
						v-model="filterQuery"
						autocomplete="off"
						autocorrect="off"
						autocapitalize="off"
						spellcheck="false"
						class="input"
						:class="{'has-autocomplete-results': autocompleteResults.length > 0}"
						:placeholder="$t('filters.query.placeholder')"
						@input="handleFieldInput"
						@focus="onFocusField"
						@keydown="onKeydown"
						@keydown.enter.prevent="blurDebounced"
						@blur="blurDebounced"
					/>
					<div
						class="filter-input-highlight"
						:style="{'height': height}"
						v-html="highlightedFilterQuery"
					/>
					<DatepickerWithValues
						v-model="currentDatepickerValue"
						v-model:open="datePickerPopupOpen"
						@update:modelValue="updateDateInQuery"
					/>
				</div>
			</template>
			<template
				#result="{ item }"
			>
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
	</div>
</template>

<style lang="scss">
.filter-input-highlight {

	&, button.filter-query__date_value {
		color: var(--card-color);
	}

	span {
		&.filter-query__field {
			color: var(--code-literal);
		}

		&.filter-query__operator {
			color: var(--code-keyword);
		}

		&.filter-query__join-operator {
			color: var(--code-section);
		}

		&.filter-query__date_value_placeholder {
			display: inline-block;
			color: transparent;
		}

		&.filter-query__assignee_value, &.filter-query__label_value {
			border-radius: $radius;
			background-color: var(--grey-200);
			color: var(--grey-700);
		}
	}

	button.filter-query__date_value {
		border-radius: $radius;
		position: absolute;
		margin-top: calc((0.25em - 0.125rem) * -1);
		height: 1.75rem;
		padding: 0;
		border: 0;
		background: transparent;
		font-size: 1rem;
		cursor: pointer;
		line-height: 1.5;
	}
}
</style>

<style lang="scss" scoped>
.filter-input {
	position: relative;

	textarea {
		position: absolute;
		background: transparent !important;
		resize: none;
		text-fill-color: transparent;
		-webkit-text-fill-color: transparent;

		&::placeholder {
			text-fill-color: var(--input-placeholder-color);
			-webkit-text-fill-color: var(--input-placeholder-color);
		}

		&.has-autocomplete-results {
			border-radius: var(--input-radius) var(--input-radius) 0 0;
		}
	}

	.filter-input-highlight {
		background: var(--white);
		height: 2.5em;
		line-height: 1.5;
		padding: .5em .75em;
		word-break: break-word;
	}
}
</style>
