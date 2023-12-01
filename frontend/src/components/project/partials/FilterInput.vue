<script setup lang="ts">
import {computed, ref, watch} from 'vue'

const {
	modelValue,
} = defineProps<{
	modelValue: string,
}>()

const filterQuery = ref('')

watch(
	() => modelValue,
	() => {
		filterQuery.value = modelValue
	},
	{immediate: true},
)

const availableFilterFields = [
	'done',
	'dueDate',
	'priority',
	'usePriority',
	'startDate',
	'endDate',
	'percentDone',
	'reminders',
	'assignees',
	'labels',
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
			></textarea>
			<div class="filter-input-highlight" v-html="highlightedFilterQuery"></div>
		</div>
	</div>
</template>

<style lang="scss">
.filter-input-highlight span {	
	&.filter-query__field {
		color: #faf594;
	}

	&.filter-query__operator {
		color: hsla(var(--primary-h), var(--primary-s), 80%);
	}

	&.filter-query__join-operator {
		color: hsla(var(--primary-h), var(--primary-s), 90%);
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
	}
	
	.filter-input-highlight {
		height: 2.5em;
		line-height: 1.5;
		padding: .5em .75em;
	}
}
</style>
