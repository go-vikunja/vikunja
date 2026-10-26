<template>
	<div
		v-if="hasActiveTimer"
		v-cy="'timerBadge'"
		class="timer-badge"
	>
		<RouterLink
			:to="{ name: 'time-tracking' }"
			class="timer-badge__elapsed"
			:title="$t('timeTracking.title')"
		>
			{{ elapsed }}
		</RouterLink>
		<BaseButton
			v-tooltip="$t('timeTracking.stop')"
			v-cy="'stopTimer'"
			class="timer-badge__stop"
			:aria-label="$t('timeTracking.stop')"
			@click="stop"
		>
			<Icon icon="stop" />
		</BaseButton>
	</div>
</template>

<script setup lang="ts">
import {ref} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import {useElapsed, useTimeTracking} from '@/composables/useTimeTracking'
import {useStopTimerMutation} from '@/client/queries/timeEntries'

const {activeTimer, hasActiveTimer} = useTimeTracking()
const stopMutation = useStopTimerMutation()
const elapsed = useElapsed(() => activeTimer.value?.start_time)

const isStopping = ref(false)
async function stop() {
	if (isStopping.value) {
		return
	}
	isStopping.value = true
	try {
		await stopMutation.mutateAsync(undefined)
	} catch {
		return
	} finally {
		isStopping.value = false
	}
}
</script>

<style lang="scss" scoped>
.timer-badge {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	white-space: nowrap;
}

.timer-badge__elapsed {
	padding-inline: .75rem .25rem;
	color: var(--primary);
	font-variant-numeric: tabular-nums;
	font-weight: 600;
}

.timer-badge__stop {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	padding-inline: .5rem;
	color: var(--grey-400);
	transition: color $transition;

	&:hover {
		color: var(--danger);
	}
}
</style>
