<template>
	<div id="filter-quick-keywords" class="keywords mbe-2">
		<div class="keywords-section">
			<span class="keywords-label">{{ $t('filters.keywords.fields') }}</span>
			<div class="keywords-buttons">
				<BaseButton
					v-for="field in filterFields"
					:key="field"
					v-tooltip="{
						content: $t(`filters.query.help.fields.${field}`),
						container: '#filter-quick-keywords',
					}"
					size="small"
					variant="ghost"
					class="keyword-btn"
					@click.prevent.stop="$emit('insert', field + ' = ')"
				>
					{{ field }}
				</BaseButton>
			</div>
		</div>
		<div class="keywords-section">
			<span class="keywords-label">{{ $t('filters.keywords.operators') }}</span>
			<div class="keywords-buttons">
				<BaseButton
					v-for="op in filterOperators"
					:key="op"
					v-tooltip="{
						content: $t(`filters.query.help.operators.${operatorKey(op)}`),
						container: '#filter-quick-keywords',
					}"
					size="small"
					variant="ghost"
					class="keyword-btn"
					@click.prevent.stop="$emit('insert', ' ' + op + ' ')"
				>
					{{ op }}
				</BaseButton>
			</div>
		</div>
		<div class="keywords-section">
			<span class="keywords-label">{{ $t('filters.keywords.join') }}</span>
			<div class="keywords-buttons">
				<BaseButton
					v-for="join in filterJoinOperators"
					:key="join"
					v-tooltip="{
						content: $t(`filters.query.help.logicalOperators.${joinOperatorKey(join)}`),
						container: '#filter-quick-keywords',
					}"
					size="small"
					variant="ghost"
					class="keyword-btn"
					@click.prevent.stop="$emit('insert', join + ' ')"
				>
					{{ join }}
				</BaseButton>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/base/BaseButton.vue'
import {
	AVAILABLE_FILTER_FIELDS,
	FILTER_OPERATORS,
	FILTER_JOIN_OPERATOR,
} from '@/helpers/filters'

defineEmits<{
	insert: [text: string],
}>()

const filterFields = AVAILABLE_FILTER_FIELDS
const filterOperators = FILTER_OPERATORS
const filterJoinOperators = FILTER_JOIN_OPERATOR

// Map operator symbols to i18n keys
function operatorKey(op: string): string {
	const mapping: Record<string, string> = {
		'!=': 'notEqual',
		'=': 'equal',
		'>': 'greaterThan',
		'>=': 'greaterThanOrEqual',
		'<': 'lessThan',
		'<=': 'lessThanOrEqual',
		like: 'like',
		'not in': 'notIn',
		in: 'in',
		'?=': 'equal',
	}
	return mapping[op] || 'equal'
}

// Map join operators to i18n keys
function joinOperatorKey(join: string): string {
	const mapping: Record<string, string> = {
		'&&': 'and',
		'||': 'or',
		'(': 'parentheses',
		')': 'parentheses',
	}
	return mapping[join] || 'and'
}
</script>

<style lang="scss" scoped>
.keywords {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	padding: 0.5rem;
	background: var(--grey-100);
	border-radius: var(--radius);
}

.keywords-section {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
}

.keywords-label {
	font-size: 0.75rem;
	font-weight: 600;
	color: var(--grey-600);
	text-transform: uppercase;
	letter-spacing: 0.05em;
}

.keywords-buttons {
	display: flex;
	flex-wrap: wrap;
	gap: 0.25rem;
}

.keyword-btn {
	font-family: monospace;
	font-size: 0.75rem;
	padding: 0.25rem 0.5rem;
	background: var(--white);
	border: 1px solid var(--grey-300);
	border-radius: var(--radius);
	transition: all var(--transition);

	&:hover {
		background: var(--primary-light);
		border-color: var(--primary);
		color: var(--primary);
	}
}
</style>
