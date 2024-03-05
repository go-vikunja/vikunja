<script setup lang="ts">
import {nextTick, ref, watch} from 'vue'
import {useAutoHeightTextarea} from '@/composables/useAutoHeightTextarea'
import DatepickerWithValues from '@/components/date/datepickerWithValues.vue'
import UserService from '@/services/user'
import {getAvatarUrl, getDisplayName} from '@/models/user'
import {createRandomID} from '@/helpers/randomId'

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
const dateFieldsRegex = '(' + dateFields.join('|') + ')'

const assigneeFields = [
	'assignees',
]

const availableFilterFields = [
	'done',
	'priority',
	'usePriority',
	'percentDone',
	'labels',
	...dateFields,
	...assigneeFields,
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

const TOKEN_REGEX = '(&lt;|&gt;|&lt;=|&gt;=|=|!=)'

function getHighlightedFilterQuery() {
	let highlighted = escapeHtml(filterQuery.value)
	dateFields
		.forEach(o => {
			const pattern = new RegExp(o + '\\s*' + TOKEN_REGEX + '\\s*([\'"]?)([^\'"\\s]+\\1?)?', 'ig')
			highlighted = highlighted.replaceAll(pattern, (match, token, start, value, position) => {
				if (typeof value === 'undefined') {
					value = ' '
				}

				return `${o} ${token} <button class="button is-primary filter-query__date_value" data-position="${position}">${value}</button>`
			})
		})
	assigneeFields
		.forEach(f => {
			const pattern = new RegExp(f + '\\s*' + TOKEN_REGEX + '\\s*([\'"]?)([^\'"\\s]+\\1?)?', 'ig')
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
}

const currentOldDatepickerValue = ref('')
const currentDatepickerValue = ref('')
const currentDatepickerPos = ref()
const datePickerPopupOpen = ref(false)

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
	updateQueryHighlight()
}

function getCharacterOffsetWithin(element: HTMLInputElement, isStart: boolean): number {
	let range = document.createRange()
	let sel = window.getSelection()
	if (sel.rangeCount > 0) {
		let originalRange = sel.getRangeAt(0)
		range.selectNodeContents(element)
		range.setEnd(
			isStart ? originalRange.startContainer : originalRange.endContainer,
			isStart ? originalRange.startOffset : originalRange.endOffset,
		)

		const rangeLength = range.toString().length
		const originalLength = originalRange.toString().length
		
		return rangeLength - (isStart ? 0 : originalLength)
	}
	return 0 // No selection
}

function saveSelectionOffsets(element: HTMLInputElement) {
	return {
		start: getCharacterOffsetWithin(element, true),
		end: getCharacterOffsetWithin(element, false),
	}
}

function setSelectionByCharacterOffsets(element: HTMLElement, startOffset: number, endOffset: number) {
	let charIndex = 0, range = document.createRange()
	const sel = window.getSelection()

	console.log({startOffset, endOffset})

	range.setStart(element, 0)
	range.collapse(true)

	let foundStart = false

	const allTextNodes: ChildNode[] = []

	element.childNodes.forEach(n => {
		if (n.nodeType === Node.TEXT_NODE) {
			allTextNodes.push(n)
		}

		n.childNodes.forEach(child => {
			if (child.nodeType === Node.TEXT_NODE) {
				allTextNodes.push(child)
			}
		})
	})

	allTextNodes.forEach(node => {
		const nextCharIndex = charIndex + node.textContent.length

		let addition = node.textContent === ' ' ? 1 : 0

		if (!foundStart && startOffset >= charIndex && startOffset <= nextCharIndex) {
			range.setStart(node, startOffset - charIndex + addition)
			foundStart = true // Start position found
		}
		if (foundStart && endOffset >= charIndex && endOffset <= nextCharIndex) {
			if (node.parentNode?.nodeName === 'BUTTON') {
				node.parentNode?.focus()
				range.setStartAfter(node.parentNode)
				range.setEndAfter(node.parentNode)
				return
			}

			range.setEnd(node, endOffset - charIndex + addition)
		}
		charIndex = nextCharIndex // Update charIndex to the next position
	})
	
	// FIXME: This kind of works for the first literal but breaks as soon as you type another query after the first it breaks

	sel.removeAllRanges()
	sel.addRange(range)
}


function updateQueryStringFromInput(e) {
	filterQuery.value = e.target.innerText
	const element = e.target

	const offsets = saveSelectionOffsets(element)
	if (offsets) {
		updateQueryHighlight()
		setSelectionByCharacterOffsets(element, offsets.start, offsets.end)
	} else {
		updateQueryHighlight()
	}
}

const queryInputRef = ref<HTMLInputElement | null>(null)

function updateQueryHighlight() {
	// Updating the query value in a function instead of a computed gives us more control about the timing
	queryInputRef.value.innerHTML = getHighlightedFilterQuery()
	nextTick(() => {
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
	})
}
</script>

<template>
	<div class="field">
		<label class="label">{{ $t('filters.query.title') }}</label>
		<div class="control filter-input">
			<div
				class="input filter-input-highlight"
				:style="{'height': height}"
				contenteditable="true"
				@input="updateQueryStringFromInput"
				ref="queryInputRef"
			></div>
			<DatepickerWithValues
				v-model="currentDatepickerValue"
				:open="datePickerPopupOpen"
				@close="() => datePickerPopupOpen = false"
				@update:model-value="updateDateInQuery"
			/>
		</div>
		{{ filterQuery }}
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
			padding: .125rem .25rem;
			display: inline-block;
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
		padding: .125rem .25rem;
		border-radius: $radius;
		margin-top: calc((0.25em - 0.125rem) * -1);
		height: 1.75rem;
	}
}
</style>

<style lang="scss" scoped>
.filter-input {
	//position: relative;

	textarea {
		//position: absolute;
		//text-fill-color: transparent;
		//-webkit-text-fill-color: transparent;
		//background: transparent !important;
		//resize: none;
	}

	.filter-input-highlight {
		height: 2.5em;
		line-height: 1.5;
		padding: .5em .75em;
		word-break: break-word;
	}
}
</style>
