<template>
	<template v-if="mode !== 'disabled' && prefixes !== undefined">
		<BaseButton
			v-tooltip="$t('task.quickAddMagic.hint')"
			class="icon is-small show-helper-text quick-add-magic-trigger-btn"
			:aria-label="$t('task.quickAddMagic.hint')"
			:class="{'is-highlighted': highlightHintIcon}"
			@click="() => visible = true"
		>
			<Icon :icon="['far', 'circle-question']" />
		</BaseButton>
		<Modal
			:enabled="visible"
			transition-name="fade"
			:overflow="true"
			variant="hint-modal"
			@close="() => visible = false"
		>
			<Card
				class="has-no-shadow"
				:title="$t('task.quickAddMagic.title')"
			>
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

				<h3>{{ $t('quickActions.projects') }}</h3>
				<p>
					{{ $t('task.quickAddMagic.project1', {prefix: prefixes.project}) }}
					{{ $t('task.quickAddMagic.project2') }}
				</p>
				<p>
					{{ $t('task.quickAddMagic.project3') }}
					{{ $t('task.quickAddMagic.project4', {prefix: prefixes.project}) }}
				</p>

				<h3>{{ $t('task.quickAddMagic.dateAndTime') }}</h3>
				<p>
					{{ $t('task.quickAddMagic.date') }}
				</p>
				<ul>
					<!-- Not localized because these only work in english -->
					<li>Today</li>
					<li>Tonight</li>
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
					<li>02/17/2021</li>
					<li>2021-02-17</li>
					<li>17.02.2021</li>
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
				</ul>
			</Card>
		</Modal>
	</template>
</template>

<script setup lang="ts">
import {ref, computed} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'

import {PREFIXES} from '@/modules/parseTaskText'
import {useAuthStore} from '@/stores/auth'

defineProps<{
	highlightHintIcon?: boolean,
}>()

const authStore = useAuthStore()

const visible = ref(false)
const mode = computed(() => authStore.settings.frontendSettings.quickAddMagicMode)

const prefixes = computed(() => PREFIXES[mode.value])
</script>

<style lang="scss" scoped>
.show-helper-text {
	// Bulma adds pointer-events: none to the icon so we need to override it back here.
	pointer-events: auto !important;
}

.is-highlighted {
	color: inherit !important;
}
</style>
