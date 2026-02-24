<template>
	<div class="gantt-grid-lines" ref="gridEl">
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

		<!-- Today column highlight: visual overlay -->
		<div
			v-if="todayIndex >= 0"
			class="today-overlay"
			:style="{
				left: (todayIndex * dayWidthPixels) + 'px',
				width: dayWidthPixels + 'px',
				height: height + 'px',
				backgroundColor: highlightColor,
			}"
		/>
	</div>

	<!-- Teleported picker button + popover: escapes z-index stacking -->
	<Teleport to="body">
		<div
			v-if="todayIndex >= 0 && btnPos"
			class="today-picker-teleported"
			:style="{
				left: btnPos.x + 'px',
				top: btnPos.y + 'px',
			}"
		>
			<button
				class="today-color-btn"
				title="Change today highlight color"
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
	</Teleport>
</template>

<script setup lang="ts">
import {computed, ref, onMounted, onUnmounted, watch, nextTick} from 'vue'
import {useStorage} from '@vueuse/core'

const props = defineProps<{
	timelineData: Date[]
	totalWidth: number
	height: number
	dayWidthPixels: number
}>()

const gridEl = ref<HTMLElement | null>(null)

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

// Compute the button's absolute position on the page
const btnPos = ref<{ x: number; y: number } | null>(null)

function updateBtnPos() {
	if (!gridEl.value || todayIndex.value < 0) {
		btnPos.value = null
		return
	}
	const rect = gridEl.value.getBoundingClientRect()
	// Position at the bottom-center of the today column
	const colLeft = todayIndex.value * props.dayWidthPixels
	const x = rect.left + colLeft + (props.dayWidthPixels / 2) - 8 // 8 = half btn width
	const y = rect.top + props.height - 28 // 28px up from bottom
	btnPos.value = { x: x + window.scrollX, y: y + window.scrollY }
}

// Update position on relevant changes
watch([() => todayIndex.value, () => props.height, () => props.dayWidthPixels], () => {
	nextTick(updateBtnPos)
})

let rafId = 0
let scrollParent: HTMLElement | null = null

function onScrollOrResize() {
	cancelAnimationFrame(rafId)
	rafId = requestAnimationFrame(updateBtnPos)
}

onMounted(() => {
	updateBtnPos()
	window.addEventListener('resize', onScrollOrResize)
	// Find the scrollable gantt container and track its scroll
	scrollParent = gridEl.value?.closest('.gantt-container') as HTMLElement | null
	if (scrollParent) {
		scrollParent.addEventListener('scroll', onScrollOrResize)
	}
	document.addEventListener('click', onDocClick)
})

onUnmounted(() => {
	window.removeEventListener('resize', onScrollOrResize)
	if (scrollParent) {
		scrollParent.removeEventListener('scroll', onScrollOrResize)
	}
	cancelAnimationFrame(rafId)
	document.removeEventListener('click', onDocClick)
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

// Today overlay: visual only
.today-overlay {
	position: absolute;
	inset-block-start: 0;
	z-index: 4;
	pointer-events: none;
	border-inline-start: 1.5px solid var(--warning);
	border-inline-end: 1.5px solid var(--warning);
}
</style>

<!-- Unscoped styles for teleported elements -->
<style lang="scss">
.today-picker-teleported {
	position: absolute;
	z-index: 9999;
	pointer-events: auto;
}

.today-color-btn {
	background: var(--grey-800, #2d2d2d);
	border: 2px solid var(--warning, #d4af37);
	border-radius: 50%;
	padding: 0;
	cursor: pointer;
	width: 16px;
	height: 16px;
	display: flex;
	align-items: center;
	justify-content: center;
	opacity: 0.6;
	transition: transform 0.15s, opacity 0.15s;

	&:hover {
		transform: scale(1.3);
		opacity: 1;
	}
}

.today-color-dot {
	display: block;
	width: 10px;
	height: 10px;
	border-radius: 50%;
}

.today-color-popover {
	position: absolute;
	bottom: 24px;
	left: calc(100% + 8px);
	background: var(--grey-800, #2d2d2d);
	border: 1px solid var(--grey-600, #555);
	border-radius: 8px;
	padding: 10px 12px;
	z-index: 10000;
	min-width: 160px;
	box-shadow: 0 4px 16px rgba(0, 0, 0, .4);
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.popover-label {
	font-size: 11px;
	font-weight: 600;
	color: var(--grey-300, #aaa);
	text-transform: uppercase;
	letter-spacing: .5px;
}

.popover-presets {
	display: flex;
	flex-wrap: wrap;
	gap: 5px;
}

.preset-swatch {
	width: 22px;
	height: 22px;
	border-radius: 4px;
	border: 2px solid transparent;
	cursor: pointer;
	padding: 0;
	transition: border-color 0.15s, transform 0.1s;

	&:hover {
		transform: scale(1.15);
	}

	&.is-active {
		border-color: #fff;
	}
}

.popover-custom {
	display: flex;
	align-items: center;
	gap: 8px;
}

.popover-custom-label {
	font-size: 11px;
	color: var(--grey-400, #888);
	white-space: nowrap;
}

.popover-color-input {
	width: 28px;
	height: 22px;
	border: none;
	padding: 0;
	cursor: pointer;
	background: transparent;
	border-radius: 3px;

	&::-webkit-color-swatch-wrapper {
		padding: 0;
	}

	&::-webkit-color-swatch {
		border: 1px solid var(--grey-500, #666);
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
	color: var(--grey-400, #888);
	white-space: nowrap;
}

.popover-opacity-slider {
	flex: 1;
	cursor: pointer;
	accent-color: var(--warning, #d4af37);
}

.popover-opacity-val {
	font-size: 11px;
	color: var(--grey-400, #888);
	min-width: 28px;
	text-align: end;
}
</style>
