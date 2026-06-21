<template>
	<div
		v-if="timeTrackingStore.hasActiveTimer"
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
import {ref, computed, onMounted, onUnmounted} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import {useTimeTrackingStore} from '@/stores/timeTracking'
import {useConfigStore} from '@/stores/config'
import {PRO_FEATURE} from '@/constants/proFeatures'

const timeTrackingStore = useTimeTrackingStore()
const configStore = useConfigStore()

const now = ref(new Date())
let interval: ReturnType<typeof setInterval> | undefined

const elapsed = computed(() => {
	const timer = timeTrackingStore.activeTimer
	if (timer === null) {
		return ''
	}
	const seconds = Math.max(0, Math.floor((now.value.getTime() - timer.startTime.getTime()) / 1000))
	const pad = (n: number) => n.toString().padStart(2, '0')
	const hours = Math.floor(seconds / 3600)
	const mmss = `${pad(Math.floor((seconds % 3600) / 60))}:${pad(seconds % 60)}`
	return hours >= 1 ? `${hours}:${mmss}` : mmss
})

const isStopping = ref(false)
async function stop() {
	if (isStopping.value) {
		return
	}
	isStopping.value = true
	try {
		await timeTrackingStore.stopTimer()
	} finally {
		isStopping.value = false
	}
}

onMounted(() => {
	// The badge lives in the always-mounted header, so it owns the app-wide timer
	// sync. Subscribing is harmless when the feature is off (no events are emitted);
	// only the hydrate hits the gated endpoint, so guard that.
	timeTrackingStore.subscribeToTimerEvents()
	if (configStore.isProFeatureEnabled(PRO_FEATURE.TIME_TRACKING)) {
		timeTrackingStore.hydrateActiveTimer()
	}
	interval = setInterval(() => {
		now.value = new Date()
	}, 1000)
})

onUnmounted(() => {
	timeTrackingStore.unsubscribeFromTimerEvents()
	if (interval !== undefined) {
		clearInterval(interval)
	}
})
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
