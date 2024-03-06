<script setup lang="ts">
import {computed, nextTick, ref, watch} from 'vue'
import {useAutoHeightTextarea} from '@/composables/useAutoHeightTextarea'
import DatepickerWithValues from '@/components/date/datepickerWithValues.vue'
import UserService from '@/services/user'
import {getAvatarUrl, getDisplayName} from '@/models/user'
import {createRandomID} from '@/helpers/randomId'
import AutocompleteDropdown from '@/components/input/AutocompleteDropdown.vue'

const {
	modelValue,
} = defineProps<{
	modelValue: string,
}>()

const filterQuery = ref('')
const {
	textarea: filterInput,
	height,
} = useAutoHeightTextarea(filterQuery)

watch(
	() => modelValue,
	() => {
		filterQuery.value = modelValue
	},
	{immediate: true},
)

const userService = new UserService()

const dateFields = [
	'dueDate',
	'startDate',
	'endDate',
	'doneAt',
	'reminders',
]

const assigneeFields = [
	'assignees',
]

const labelFields = [
	'labels',
]

const availableFilterFields = [
	'done',
	'priority',
	'usePriority',
	'percentDone',
	...dateFields,
	...assigneeFields,
	...labelFields,
]

const filterOperators = [
	'!=',
	'=',
	'>',
	'>=',
	'<',
	'<=',
	'like',
	'in',
	'?=',
]

const filterJoinOperators = [
	'&&',
	'||',
	'(',
	')',
]

const FILTER_OPERATORS_REGEX = '(&lt;|&gt;|&lt;=|&gt;=|=|!=)'
const FILTER_JOIN_OPERATORS_REGEX = '(&amp;&amp;|\|\||\(|\))'

function escapeHtml(unsafe: string): string {
	return unsafe
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#039;')
}

function unEscapeHtml(unsafe: string): string {
	return unsafe
		.replace(/&amp;/g, '&')
		.replace(/&lt;/g, '<')
		.replace(/&gt;/g, '>')
		.replace(/&quot/g, '"')
		.replace(/&#039;/g, '\'')
}

const highlightedFilterQuery = computed(() => {
	let highlighted = escapeHtml(filterQuery.value)
	dateFields
		.forEach(o => {
			const pattern = new RegExp(o + '(\\s*)' + FILTER_OPERATORS_REGEX + '(\\s*)([\'"]?)([^\'"\\s]+\\1?)?', 'ig')
			highlighted = highlighted.replaceAll(pattern, (match, spacesBefore, token, spacesAfter, start, value, position) => {
				if (typeof value === 'undefined') {
					value = ''
				}

				return `${o}${spacesBefore}${token}${spacesAfter}<button class="is-primary filter-query__date_value" data-position="${position}">${value}</button><span class="filter-query__date_value_placeholder">${value}</span>`
			})
		})
	assigneeFields
		.forEach(f => {
			const pattern = new RegExp(f + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*([\'"]?)([^\'"\\s]+\\1?)?', 'ig')
			highlighted = highlighted.replaceAll(pattern, (match, token, start, value) => {
				if (typeof value === 'undefined') {
					value = ''
				}

				const id = createRandomID(32)

				userService.getAll({}, {s: value}).then(users => {
					if (users.length > 0) {
						const displayName = getDisplayName(users[0])
						const nameTag = document.createElement('span')
						nameTag.innerText = displayName

						const avatar = document.createElement('img')
						avatar.src = getAvatarUrl(users[0], 20)
						avatar.height = 20
						avatar.width = 20
						avatar.alt = displayName

						// TODO: caching

						nextTick(() => {
							const assigneeValue = document.getElementById(id)
							assigneeValue.innerText = ''
							assigneeValue?.appendChild(avatar)
							assigneeValue?.appendChild(nameTag)
						})
					}
				})

				return `${f} ${token} <span class="filter-query__assignee_value" id="${id}">${value}<span>`
			})
		})
	filterOperators
		.map(o => ` ${escapeHtml(o)} `)
		.forEach(o => {
			highlighted = highlighted.replaceAll(o, `<span class="filter-query__operator">${o}</span>`)
		})
	filterJoinOperators
		.map(o => escapeHtml(o))
		.forEach(o => {
			highlighted = highlighted.replaceAll(o, `<span class="filter-query__join-operator">${o}</span>`)
		})
	availableFilterFields.forEach(f => {
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

function handleFieldInput(e, autocompleteOnInput) {
	const cursorPosition = filterInput.value.selectionStart
	const textUpToCursor = filterQuery.value.substring(0, cursorPosition)

	labelFields.forEach(l => {
		const pattern = new RegExp('(' + l + '\\s*' + FILTER_OPERATORS_REGEX + '\\s*)([\'"]?)([^\'"&\|\(\)]+\\1?)?$', 'ig')
		const match = pattern.exec(textUpToCursor)

		if (match !== null) {
			const [matched, prefix, operator, space, keyword] = match
			if (keyword) {
				autocompleteResults.value = ['loool', keyword]
			}
		}
	})
}

const autocompleteResults = ref<any[]>([])
</script>

<template>
	<div class="field">
		<label class="label">{{ $t('filters.query.title') }}</label>
		<AutocompleteDropdown 
			:options="autocompleteResults"
			@blur="filterInput?.blur()"
		>
			<template
				v-slot:input="{ onKeydown, onFocusField, onUpdateField }"
			>
				<div class="control filter-input">
					<textarea
						@input="e => handleFieldInput(e, onUpdateField)"
						@focus="onFocusField"
						@keydown="onKeydown"

						autocomplete="off"
						autocorrect="off"
						autocapitalize="off"
						spellcheck="false"
						v-model="filterQuery"
						class="input"
						ref="filterInput"
					></textarea>
					<div
						class="filter-input-highlight"
						:style="{'height': height}"
						v-html="highlightedFilterQuery"
					></div>
					<DatepickerWithValues
						v-model="currentDatepickerValue"
						:open="datePickerPopupOpen"
						@close="() => datePickerPopupOpen = false"
						@update:model-value="updateDateInQuery"
					/>
				</div>
			</template>
			<template
				v-slot:result="{ item }"
			>
				whoo {{ item }}
			</template>
		</AutocompleteDropdown>
	</div>
</template>

<style lang="scss">
.filter-input-highlight {
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

		&.filter-query__assignee_value {
			padding: .125rem .25rem;
			border-radius: $radius;
			background-color: var(--grey-200);
			color: var(--grey-700);
			display: inline-flex;
			align-items: center;

			> img {
				margin-right: .25rem;
			}
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
		text-fill-color: transparent;
		-webkit-text-fill-color: transparent;
		background: transparent !important;
		resize: none;
	}

	.filter-input-highlight {
		height: 2.5em;
		line-height: 1.5;
		padding: .5em .75em;
		word-break: break-word;
	}
}
</style>
