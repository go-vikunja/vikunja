<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useAutoHeightTextarea} from '@/composables/useAutoHeightTextarea'

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

const dateFields = [
	'dueDate',
	'startDate',
	'endDate',
	'doneAt',
]

const availableFilterFields = [
	'done',
	'priority',
	'usePriority',
	'percentDone',
	'reminders',
	'assignees',
	'labels',
	...dateFields,
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

const highlightedFilterQuery = computed(() => {
	let highlighted = escapeHtml(filterQuery.value)
	dateFields
		.map(o => escapeHtml(o))
		.forEach(o => {
			const pattern = new RegExp(o + '\\s*(&lt;|&gt;|&lt;=|&gt;=|=|!=)\\s*([\'"]?)([^\'"\\s]+\\1?)?', 'ig');
			highlighted = highlighted.replaceAll(pattern, (match, token, start, value, position) => {
				console.log({match, token, value, position})
				return `${o} ${token} <span class="filter-query__special_value">${value}</span><span class="filter-query__special_value_placeholder">${value}</span>`
				// TODO: make special value a button with datepicker popup
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

function escapeHtml(unsafe: string): string {
	return unsafe
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#039;')
}
</script>

<template>
	<div class="field">
		<label class="label">{{ $t('filters.query.title') }}</label>
		<div class="control filter-input">
			<textarea
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
		</div>
	</div>
</template>

<style lang="scss">
.filter-input-highlight span {
	&.filter-query__field {
		color: var(--code-literal);
	}

	&.filter-query__operator {
		color: var(--code-keyword);
	}

	&.filter-query__join-operator {
		color: var(--code-section);
	}

	&.filter-query__special_value_placeholder {
		padding: .125rem .25rem;
		display: inline-block;
	}
	
	&.filter-query__special_value {
		background: var(--primary);
		padding: .125rem .25rem;
		color: white;
		border-radius: $radius;
		position: absolute;
		margin-top: calc((0.25em - 0.125rem) * -1);
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
