<template>
	<div v-if="available">
		<p class="help has-text-grey">
			{{ $t('task.quickAddMagic.hint') }}.
			<a @click="() => visible = true">{{ $t('task.quickAddMagic.what') }}</a>
		</p>
		<modal
			@close="() => visible = false"
			:enabled="visible"
			transition-name="fade"
			:overflow="true"
			variant="hint-modal"
		>
			<card class="has-no-shadow" :title="$t('task.quickAddMagic.title')">
				<p>{{ $t('task.quickAddMagic.intro') }}</p>

				<h3>{{ $t('task.attributes.labels') }}</h3>
				<p>
					{{ $t('task.quickAddMagic.label1', {prefix: prefixes.label}) }}
					{{ $t('task.quickAddMagic.label2') }}
					{{ $t('task.quickAddMagic.multiple') }}
				</p>
				<p>
					{{ $t('task.quickAddMagic.label3') }}
					{{ $t('task.quickAddMagic.label4', {prefix: prefixes.label}) }}
				</p>

				<h3>{{ $t('task.attributes.priority') }}</h3>
				<p>
					{{ $t('task.quickAddMagic.priority1', {prefix: prefixes.priority}) }}
					{{ $t('task.quickAddMagic.priority2') }}
				</p>

				<h3>{{ $t('task.attributes.assignees') }}</h3>
				<p>
					{{ $t('task.quickAddMagic.assignees', {prefix: prefixes.assignee}) }}
					{{ $t('task.quickAddMagic.multiple') }}
				</p>

				<h3>{{ $t('list.list.title') }}</h3>
				<p>
					{{ $t('task.quickAddMagic.list1', {prefix: prefixes.list}) }}
					{{ $t('task.quickAddMagic.list2') }}
				</p>

				<h3>{{ $t('task.quickAddMagic.dateAndTime') }}</h3>
				<p>
					{{ $t('task.quickAddMagic.date') }}
				</p>
				<ul>
					<!-- Not localized because these only work in english -->
					<li>Today</li>
					<li>Tomorrow</li>
					<li>Next monday</li>
					<li>This weekend</li>
					<li>Later this week</li>
					<li>Later next week</li>
					<li>Next week</li>
					<li>Next month</li>
					<li>End of month</li>
					<li>In 5 days [hours/weeks/months]</li>
					<li>Tuesday ({{ $t('task.quickAddMagic.dateWeekday') }})</li>
					<li>17/02/2021</li>
					<li>Feb 17 ({{ $t('task.quickAddMagic.dateCurrentYear') }})</li>
					<li>17th ({{ $t('task.quickAddMagic.dateNth', {day: '17'}) }})</li>
				</ul>
				<p>{{ $t('task.quickAddMagic.dateTime', {time: 'at 17:00', timePM: '5pm'}) }}</p>
				
				<h3>{{ $t('task.quickAddMagic.repeats') }}</h3>
				<p>{{ $t('task.quickAddMagic.repeatsDescription', {suffix: 'every {amount} {type}'}) }}</p>
				<p>{{ $t('misc.forExample') }}</p>
				<ul>
					<!-- Not localized because these only work in english -->
					<li>Every day</li>
					<li>Every 3 days</li>
					<li>Every week</li>
					<li>Every 2 weeks</li>
					<li>Every month</li>
					<li>Every 6 months</li>
					<li>Every year</li>
					<li>Every 2 years</li>
				</ul>
			</card>
		</modal>
	</div>
</template>

<script>
import {getQuickAddMagicMode} from '@/helpers/quickAddMagicMode'
import {PREFIXES} from '@/modules/parseTaskText'

export default {
	name: 'quick-add-magic',
	data() {
		return {
			visible: false,
			mode: getQuickAddMagicMode(),
		}
	},
	computed: {
		available() {
			return this.mode !== 'disabled'
		},
		prefixes() {
			return PREFIXES[this.mode]
		},
	},
}
</script>
