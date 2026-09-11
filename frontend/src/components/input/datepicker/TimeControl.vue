<template>
	<div class="time-control">
		<div class="time-control__clock">
			<Icon
				:icon="['far', 'clock']"
				class="time-control__icon"
			/>
			<div class="time-control__segment">
				<button
					v-tooltip="$t('input.datepicker.time.hourUp')"
					type="button"
					class="time-control__step"
					:aria-label="$t('input.datepicker.time.hourUp')"
					@click.stop="step('hours', 1)"
				>
					<Icon icon="chevron-up" />
				</button>
				<input
					class="time-control__value"
					type="text"
					inputmode="numeric"
					:value="displayHours"
					:aria-label="$t('input.datepicker.time.hours')"
					@change="onInput('hours', $event)"
					@keydown.up.prevent="step('hours', 1)"
					@keydown.down.prevent="step('hours', -1)"
				>
				<button
					v-tooltip="$t('input.datepicker.time.hourDown')"
					type="button"
					class="time-control__step"
					:aria-label="$t('input.datepicker.time.hourDown')"
					@click.stop="step('hours', -1)"
				>
					<Icon icon="chevron-down" />
				</button>
			</div>
			<span class="time-control__colon">:</span>
			<div class="time-control__segment">
				<button
					v-tooltip="$t('input.datepicker.time.minuteUp')"
					type="button"
					class="time-control__step"
					:aria-label="$t('input.datepicker.time.minuteUp')"
					@click.stop="step('minutes', 5)"
				>
					<Icon icon="chevron-up" />
				</button>
				<input
					class="time-control__value"
					type="text"
					inputmode="numeric"
					:value="pad(minutes)"
					:aria-label="$t('input.datepicker.time.minutes')"
					@change="onInput('minutes', $event)"
					@keydown.up.prevent="step('minutes', 1)"
					@keydown.down.prevent="step('minutes', -1)"
				>
				<button
					v-tooltip="$t('input.datepicker.time.minuteDown')"
					type="button"
					class="time-control__step"
					:aria-label="$t('input.datepicker.time.minuteDown')"
					@click.stop="step('minutes', -5)"
				>
					<Icon icon="chevron-down" />
				</button>
			</div>
			<button
				v-if="!is24h"
				type="button"
				class="time-control__meridiem"
				@click.stop="toggleMeridiem"
			>
				{{ meridiemLabel }}
			</button>
		</div>

		<div class="time-control__presets">
			<button
				v-for="preset in PRESETS"
				:key="preset"
				type="button"
				class="time-control__preset"
				:class="{'is-active': hours === preset && minutes === 0}"
				@click.stop="emit('update', {hours: preset, minutes: 0})"
			>
				{{ formatPreset(preset) }}
			</button>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick} from 'vue'

import {useTimeFormat} from '@/composables/useTimeFormat'
import {TIME_FORMAT} from '@/constants/timeFormat'
import {formatDate} from '@/helpers/time/formatDate'

const props = defineProps<{
	hours: number
	minutes: number
}>()

const emit = defineEmits<{
	update: [value: {hours: number, minutes: number}]
}>()

const PRESETS = [9, 12, 17]

const {store: timeFormat} = useTimeFormat()
const is24h = computed(() => timeFormat.value === TIME_FORMAT.HOURS_24)

const pad = (n: number) => String(n).padStart(2, '0')

const displayHours = computed(() => is24h.value
	? pad(props.hours)
	: pad(((props.hours + 11) % 12) + 1))

const hourAsDate = (hours: number) => new Date(2024, 0, 1, hours)
const meridiemLabel = computed(() => formatDate(hourAsDate(props.hours), 'A'))

function step(part: 'hours' | 'minutes', delta: number) {
	if (part === 'hours') {
		emit('update', {hours: (props.hours + delta + 24) % 24, minutes: props.minutes})
		return
	}
	const total = (((props.hours * 60 + props.minutes + delta) % 1440) + 1440) % 1440
	emit('update', {hours: Math.floor(total / 60), minutes: total % 60})
}

async function onInput(part: 'hours' | 'minutes', event: Event) {
	const target = event.target as HTMLInputElement
	const value = Number.parseInt(target.value, 10)

	if (!Number.isNaN(value)) {
		if (part === 'minutes') {
			emit('update', {hours: props.hours, minutes: Math.min(59, Math.max(0, value))})
		} else if (is24h.value) {
			emit('update', {hours: Math.min(23, Math.max(0, value)), minutes: props.minutes})
		} else {
			// Typed "12" in 12h mode means 12 AM / 12 PM, not 0/24.
			const clamped = Math.min(12, Math.max(1, value))
			emit('update', {hours: (clamped % 12) + (props.hours >= 12 ? 12 : 0), minutes: props.minutes})
		}
		await nextTick()
	}

	target.value = part === 'hours' ? displayHours.value : pad(props.minutes)
}

function toggleMeridiem() {
	emit('update', {hours: (props.hours + 12) % 24, minutes: props.minutes})
}

function formatPreset(hours: number) {
	if (is24h.value) {
		return `${pad(hours)}:00`
	}
	return formatDate(hourAsDate(hours), 'h A')
}
</script>

<style lang="scss" scoped>
.time-control {
	display: flex;
	align-items: center;
	flex-wrap: wrap;
	gap: .75rem 1rem;
	padding: .75rem 1rem;
	border-block-start: 1px solid var(--grey-200);
}

.time-control__clock {
	display: flex;
	align-items: center;
	gap: .25rem;
	color: var(--grey-500);
}

.time-control__icon {
	margin-inline-end: .25rem;
}

.time-control__segment {
	display: flex;
	flex-direction: column;
	align-items: center;
}

.time-control__step {
	border: 0;
	background: transparent;
	color: var(--grey-400);
	font-size: .6rem;
	line-height: 1;
	padding: .125rem .25rem;
	cursor: pointer;
	transition: color $transition;

	&:hover {
		color: var(--primary);
	}

	// WCAG 2.5.8 minimum target size; only the sheet is touch-first, the desktop popup stays compact.
	.bottom-sheet & {
		display: flex;
		align-items: center;
		justify-content: center;
		min-inline-size: 24px;
		min-block-size: 24px;
	}
}

.time-control__value {
	inline-size: 2.25rem;
	border: 0;
	background: transparent;
	text-align: center;
	font-family: $vikunja-font;
	font-weight: 700;
	font-size: 1.35rem;
	line-height: 1;
	color: var(--grey-800);
	font-variant-numeric: tabular-nums;
	padding: .1875rem 0;

	&:focus {
		outline: none;
		border-radius: $radius;
		@include focus-ring;
	}
}

.time-control__colon {
	font-family: $vikunja-font;
	font-weight: 700;
	font-size: 1.35rem;
	color: var(--grey-400);
	padding-block-end: .0625rem;
}

.time-control__meridiem {
	margin-inline-start: .375rem;
	padding: .25rem .5rem;
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	background: var(--white);
	color: var(--grey-600);
	font-weight: 700;
	font-size: .7rem;
	letter-spacing: .04em;
	cursor: pointer;
}

.time-control__presets {
	display: flex;
	gap: .375rem;
	margin-inline-start: auto;
}

.time-control__preset {
	padding: .25rem .5625rem;
	border: 1px solid var(--grey-200);
	border-radius: $radius-rounded;
	background: var(--white);
	color: var(--grey-600);
	font-size: .75rem;
	font-weight: 600;
	cursor: pointer;
	transition: all $transition;

	&:hover {
		border-color: var(--grey-300);
	}

	&.is-active {
		border-color: var(--primary);
		background: hsla(var(--primary-hsl), .15);
		color: var(--primary);
	}
}
</style>
