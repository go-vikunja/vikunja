<script setup lang="ts">
import {ref} from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import Expandable from '@/components/base/Expandable.vue'

const showDocs = ref(false)
</script>

<template>
	<BaseButton
		class="has-text-primary"
		@click="showDocs = !showDocs"
	>
		{{ $t('filters.query.help.link') }}
	</BaseButton>

	<Expandable
		:open="showDocs"
		class="content"
	>
		<p>{{ $t('filters.query.help.intro') }}</p>
		<ul>
			<li><code>done</code>: {{ $t('filters.query.help.fields.done') }}</li>
			<li><code>priority</code>: {{ $t('filters.query.help.fields.priority') }}</li>
			<li><code>percentDone</code>: {{ $t('filters.query.help.fields.percentDone') }}</li>
			<li><code>dueDate</code>: {{ $t('filters.query.help.fields.dueDate') }}</li>
			<li><code>startDate</code>: {{ $t('filters.query.help.fields.startDate') }}</li>
			<li><code>endDate</code>: {{ $t('filters.query.help.fields.endDate') }}</li>
			<li><code>doneAt</code>: {{ $t('filters.query.help.fields.doneAt') }}</li>
			<li><code>assignees</code>: {{ $t('filters.query.help.fields.assignees') }}</li>
			<li><code>labels</code>: {{ $t('filters.query.help.fields.labels') }}</li>
			<li><code>project</code>: {{ $t('filters.query.help.fields.project') }}</li>
			<li><code>reminders</code>: {{ $t('filters.query.help.fields.reminders') }}</li>
			<li><code>created</code>: {{ $t('filters.query.help.fields.created') }}</li>
			<li><code>updated</code>: {{ $t('filters.query.help.fields.updated') }}</li>
		</ul>
		<p>{{ $t('filters.query.help.canUseDatemath') }}</p>
		<p>{{ $t('filters.query.help.operators.intro') }}</p>
		<ul>
			<li><code>!=</code>: {{ $t('filters.query.help.operators.notEqual') }}</li>
			<li><code>=</code>: {{ $t('filters.query.help.operators.equal') }}</li>
			<li><code>&gt;</code>: {{ $t('filters.query.help.operators.greaterThan') }}</li>
			<li><code>&gt;=</code>: {{ $t('filters.query.help.operators.greaterThanOrEqual') }}</li>
			<li><code>&lt;</code>: {{ $t('filters.query.help.operators.lessThan') }}</li>
			<li><code>&lt;=</code>: {{ $t('filters.query.help.operators.lessThanOrEqual') }}</li>
			<li><code>like</code>: {{ $t('filters.query.help.operators.like') }}</li>
			<li><code>in</code>: {{ $t('filters.query.help.operators.in') }}</li>
			<li><code>not in</code>: {{ $t('filters.query.help.operators.notIn') }}</li>
		</ul>
		<p>{{ $t('filters.query.help.logicalOperators.intro') }}</p>
		<ul>
			<li><code>&amp;&amp;</code>: {{ $t('filters.query.help.logicalOperators.and') }}</li>
			<li><code>||</code>: {{ $t('filters.query.help.logicalOperators.or') }}</li>
			<li><code>(</code> and <code>)</code>: {{ $t('filters.query.help.logicalOperators.parentheses') }}</li>
		</ul>
		<p>{{ $t('filters.query.help.examples.intro') }}</p>
		<ul>
			<li><code>priority = 4</code>: {{ $t('filters.query.help.examples.priorityEqual') }}</li>
			<li><code>dueDate &lt; now</code>: {{ $t('filters.query.help.examples.dueDatePast') }}</li>
			<li>
				<code>done = false &amp;&amp; priority &gt;= 3</code>:
				{{ $t('filters.query.help.examples.undoneHighPriority') }}
			</li>
			<li><code>assignees in user1, user2</code>: {{ $t('filters.query.help.examples.assigneesIn') }}</li>
			<li>
				<code>(priority = 1 || priority = 2) &amp;&amp; dueDate &lt;= now</code>:
				{{ $t('filters.query.help.examples.priorityOneOrTwoPastDue') }}
			</li>
		</ul>
	</Expandable>
</template>
