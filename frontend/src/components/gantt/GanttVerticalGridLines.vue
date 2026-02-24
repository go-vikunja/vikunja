<template>
	<div class="gantt-grid-lines">
		<svg
			class="gantt-vertical-lines"
			:width="totalWidth"
			:height="height"
			xmlns="http://www.w3.org/2000/svg"
		>
			<line
				v-for="(date, index) in timelineData"
				:key="date.toISOString()"
				:x1="index * dayWidthPixels"
				:y1="0"
				:x2="index * dayWidthPixels"
				:y2="height"
				stroke="var(--grey-400)"
				stroke-width="0.5"
				opacity="0.6"
			/>
		</svg>

		<!-- Today column highlight: positioned ABOVE rows with mix-blend-mode -->
		<div
			v-if="todayIndex >= 0"
			class="today-overlay"
			:style="{
				left: (todayIndex * dayWidthPixels) + 'px',
				width: dayWidthPixels + 'px',
				height: height + 'px',
				backgroundColor: highlightColor,
			}"
		>
			<!-- Color picker button at top of today column -->
			<button
				class="today-color-btn"
				:title="'Change today highlight color'"
				@click.stop="showPicker = !showPicker"
			>
				<span class="today-color-dot" :style="{ backgroundColor: hexColor }" />
			</button>

			<!-- Color picker popover -->
			<div
				v-if="showPicker"
				class="today-color-popover"
				@click.stop
			>
				<div class="popover-label">Today highlight</div>
				<div class="popover-presets">
					<button
						v-for="preset in presets"
						:key="preset.hex"
						class="preset-swatch"
						:class="{ 'is-active': hexColor === preset.hex }"
						:style="{ backgroundColor: preset.hex }"
						:title="preset.name"
						@click="applyColor(preset.hex)"
					/>
				</div>
				<div class="popover-custom">
					<label class="popover-custom-label">Custom:</label>
					<input
						type="color"
						:value="hexColor"
						class="popover-color-input"
						@input="onCustomColor"
					>
				</div>
				<div class="popover-opacity">
					<label class="popover-opacity-label">Opacity:</label>
					<input
						type="range"
						min="5"
						max="40"
						:value="opacityPercent"
						class="popover-opacity-slider"
						@input="onOpacityChange"
					>
					<span class="popover-opacity-val">{{ opacityPercent }}%</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, onMounted, onUnmounted} from 'vue'
import {useStorage} from '@vueuse/core'

const props = defineProps<{
	timelineData: Date[]
	totalWidth: number
	height: number
	dayWidthPixels: number
}>()

// Persisted settings
const storedHex = useStorage('ganttTodayHex', '#d4af37')
const storedOpacity = useStorage('ganttTodayOpacity', 15)

const showPicker = ref(false)

const hexColor = computed(() => storedHex.value)
const opacityPercent = computed(() => storedOpacity.value)

const highlightColor = computed(() => {
	const hex = storedHex.value
	const r = parseInt(hex.slice(1, 3), 16)
	const g = parseInt(hex.slice(3, 5), 16)
	const b = parseInt(hex.slice(5, 7), 16)
	return `rgba(${r}, ${g}, ${b}, ${storedOpacity.value / 100})`
})

const todayIndex = computed(() => {
	const now = new Date()
	const todayStr = now.toDateString()
	return props.timelineData.findIndex(d => d.toDateString() === todayStr)
})

const presets = [
	{ hex: '#d4af37', name: 'Gold' },
	{ hex: '#f5c542', name: 'Yellow' },
	{ hex: '#ff9800', name: 'Orange' },
	{ hex: '#4caf50', name: 'Green' },
	{ hex: '#2196f3', name: 'Blue' },
	{ hex: '#9c27b0', name: 'Purple' },
	{ hex: '#ef5350', name: 'Red' },
	{ hex: '#ffffff', name: 'White' },
]

function applyColor(hex: string) {
	storedHex.value = hex
}

function onCustomColor(e: Event) {
	storedHex.value = (e.target as HTMLInputElement).value
}

function onOpacityChange(e: Event) {
	storedOpacity.value = parseInt((e.target as HTMLInputElement).value)
}

// Close picker on outside click
function onDocClick() {
	showPicker.value = false
}
onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>

<style scoped lang="scss">
.gantt-grid-lines {
	position: absolute;
	inset-inline-start: 0;
	z-index: 1;
	pointer-events: none;
}

.gantt-vertical-lines {
	position: absolute;
	inset: 0;
}

// Today overlay: sits above rows (z-index 4) with low opacity
// so project row colors remain clearly visible underneath
.today-overlay {
	position: absolute;
	inset-block-start: 0;
	z-index: 4;
	pointer-events: none;
	border-inline-start: 1.5px solid var(--warning);
	border-inline-end: 1.5px solid var(--warning);
}

// The button needs pointer events - sits at bottom of today column
.today-color-btn {
	pointer-events: auto;
	position: absolute;
	inset-block-end: 4px;
	inset-inline-start: 50%;
	transform: translateX(-50%);
	background: var(--grey-800);
	border: 1.5px solid var(--warning);
	border-radius: 50%;
	padding: 0;
	cursor: pointer;
	inline-size: 16px;
	block-size: 16px;
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 10;
	opacity: 0.5;
	transition: transform 0.15s, opacity 0.15s;

	&:hover {
		transform: translateX(-50%) scale(1.3);
		opacity: 1;
	}
}

.today-color-dot {
	display: block;
	inline-size: 8px;
	block-size: 8px;
	border-radius: 50%;
}

// Popover - opens upward from the bottom button
.today-color-popover {
	pointer-events: auto;
	position: absolute;
	inset-block-end: 24px;
	inset-inline-start: calc(100% + 8px);
	background: var(--grey-800);
	border: 1px solid var(--grey-600);
	border-radius: 8px;
	padding: 10px 12px;
	z-index: 20;
	min-inline-size: 160px;
	box-shadow: 0 4px 16px rgba(0, 0, 0, .4);
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.popover-label {
	font-size: 11px;
	font-weight: 600;
	color: var(--grey-300);
	text-transform: uppercase;
	letter-spacing: .5px;
}

.popover-presets {
	display: flex;
	flex-wrap: wrap;
	gap: 5px;
}

.preset-swatch {
	pointer-events: auto;
	inline-size: 22px;
	block-size: 22px;
	border-radius: 4px;
	border: 2px solid transparent;
	cursor: pointer;
	padding: 0;
	transition: border-color 0.15s, transform 0.1s;

	&:hover {
		transform: scale(1.15);
	}

	&.is-active {
		border-color: var(--white);
	}
}

.popover-custom {
	display: flex;
	align-items: center;
	gap: 8px;
}

.popover-custom-label {
	font-size: 11px;
	color: var(--grey-400);
	white-space: nowrap;
}

.popover-color-input {
	inline-size: 28px;
	block-size: 22px;
	border: none;
	padding: 0;
	cursor: pointer;
	background: transparent;
	border-radius: 3px;

	&::-webkit-color-swatch-wrapper {
		padding: 0;
	}

	&::-webkit-color-swatch {
		border: 1px solid var(--grey-500);
		border-radius: 3px;
	}
}

.popover-opacity {
	display: flex;
	align-items: center;
	gap: 6px;
}

.popover-opacity-label {
	font-size: 11px;
	color: var(--grey-400);
	white-space: nowrap;
}

.popover-opacity-slider {
	flex: 1;
	cursor: pointer;
	accent-color: var(--warning);
}

.popover-opacity-val {
	font-size: 11px;
	color: var(--grey-400);
	min-inline-size: 28px;
	text-align: end;
}
</style>
