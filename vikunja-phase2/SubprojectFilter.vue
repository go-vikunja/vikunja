<template>
	<div
		v-if="childProjects.length > 0"
		class="subproject-filter"
	>
		<!-- Gantt mode: checkbox + always-visible legend -->
		<template v-if="showLegend">
			<FancyCheckbox
				:model-value="includeSubprojects"
				is-block
				@update:modelValue="toggleAll"
			>
				{{ $t('task.template.includeSubprojects') }}
			</FancyCheckbox>

			<div
				v-if="includeSubprojects"
				class="subproject-legend"
			>
				<span
					v-if="parentEntry"
					class="legend-item is-parent"
				>
					<span
						class="legend-dot"
						:style="{ backgroundColor: parentEntry.color }"
					/>
					<span class="legend-label">{{ parentEntry.title }}</span>
				</span>
				<BaseButton
					v-for="child in childLegendEntries"
					:key="child.id"
					class="legend-item"
					:class="{ 'is-excluded': excludedIds.has(child.id) }"
					@click.prevent.stop="toggleProject(child.id)"
				>
					<span
						class="legend-dot"
						:style="{ backgroundColor: excludedIds.has(child.id) ? 'var(--grey-400)' : child.color }"
					/>
					<span class="legend-label">{{ child.title }}</span>
				</BaseButton>
			</div>
		</template>

		<!-- List/Table mode: simple toggle button -->
		<template v-else>
			<XButton
				:variant="includeSubprojects ? 'primary' : 'secondary'"
				icon="sitemap"
				:shadow="false"
				class="mis-2"
				@click.prevent.stop="toggleAll(!includeSubprojects)"
			>
				{{ $t('task.template.subprojects') }}
				<span
					v-if="includeSubprojects"
					class="subproject-badge"
				>
					{{ enabledCount }}/{{ childProjects.length }}
				</span>
			</XButton>
			<div
				v-if="includeSubprojects"
				class="subproject-dropdown-wrap"
			>
				<Popup>
					<template #trigger="{ toggle }">
						<BaseButton
							class="subproject-chevron"
							@click.prevent.stop="toggle()"
						>
							<Icon icon="chevron-down" />
						</BaseButton>
					</template>
					<template #content>
						<Card class="subproject-popup">
							<div
								v-for="child in childLegendEntries"
								:key="child.id"
								class="subproject-item"
							>
								<FancyCheckbox
									:model-value="!excludedIds.has(child.id)"
									@update:modelValue="toggleProject(child.id)"
								>
									<span class="subproject-label">
										<span
											class="subproject-color-dot"
											:style="{ backgroundColor: child.color }"
										/>
										{{ child.title }}
									</span>
								</FancyCheckbox>
							</div>
						</Card>
					</template>
				</Popup>
			</div>
		</template>
	</div>
</template>

<script lang="ts" setup>
import {ref, computed, watch, onMounted} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import Popup from '@/components/misc/Popup.vue'
import Card from '@/components/misc/Card.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'

import {useSubprojectColors} from '@/composables/useSubprojectColors'

import type {IProject} from '@/modelTypes/IProject'

const props = withDefaults(defineProps<{
	projectId: IProject['id']
	showLegend?: boolean
}>(), {
	showLegend: false,
})

const emit = defineEmits<{
	'update:includeSubprojects': [value: boolean]
	'update:excludeProjectIds': [value: string]
	'update:colorMap': [value: Map<number, string>]
}>()

const projectIdRef = computed(() => props.projectId)
const {childProjects, legend, colorMap} = useSubprojectColors(projectIdRef)

// Load persisted state from localStorage
function loadState() {
	try {
		const raw = localStorage.getItem(`subprojectFilter_${props.projectId}`)
		if (raw) {
			const parsed = JSON.parse(raw)
			return {enabled: !!parsed.enabled, excluded: new Set<number>(parsed.excluded || [])}
		}
	} catch { /* ignore */ }
	return {enabled: false, excluded: new Set<number>()}
}

function saveState() {
	localStorage.setItem(`subprojectFilter_${props.projectId}`, JSON.stringify({
		enabled: includeSubprojects.value,
		excluded: Array.from(excludedIds.value),
	}))
}

const initial = loadState()
const includeSubprojects = ref(initial.enabled)
const excludedIds = ref<Set<number>>(initial.excluded)

const childProjectsWithColors = computed(() => legend.value)

const parentEntry = computed(() => {
	return legend.value.length > 0 ? legend.value[0] : null
})

const childLegendEntries = computed(() => {
	return legend.value.slice(1)
})

const enabledCount = computed(() => {
	return childProjects.value.filter(c => !excludedIds.value.has(c.id)).length
})

function toggleAll(enabled: boolean) {
	includeSubprojects.value = enabled
	if (!enabled) {
		excludedIds.value = new Set()
	}
	emitUpdate()
}

function toggleProject(id: number) {
	const newSet = new Set(excludedIds.value)
	if (newSet.has(id)) {
		newSet.delete(id)
	} else {
		newSet.add(id)
	}
	excludedIds.value = newSet

	// If all child projects are now excluded, auto-disable subprojects entirely
	if (newSet.size >= childProjects.value.length && includeSubprojects.value) {
		includeSubprojects.value = false
		excludedIds.value = new Set()
	}

	emitUpdate()
}

function emitUpdate() {
	emit('update:includeSubprojects', includeSubprojects.value)
	emit('update:excludeProjectIds', Array.from(excludedIds.value).join(','))
	emit('update:colorMap', includeSubprojects.value ? colorMap.value : new Map())
	saveState()
}
// Emit persisted state immediately during setup so GanttChart has colors on first render
if (includeSubprojects.value) {
	emitUpdate()
}
// Also emit on mount in case colorMap was not yet computed during setup
onMounted(() => {
	if (includeSubprojects.value) {
		emitUpdate()
	}
})

// Reload state when navigating to a different project
watch(() => props.projectId, () => {
	const state = loadState()
	includeSubprojects.value = state.enabled
	excludedIds.value = state.excluded
	emitUpdate()
})

// Re-emit when colorMap updates (e.g., project store finishes loading)
watch(colorMap, () => {
	if (includeSubprojects.value) {
		emitUpdate()
	}
})
</script>

<style lang="scss" scoped>
.subproject-filter {
	display: inline-flex;
	align-items: center;
	gap: .75rem;
}

.subproject-badge {
	background: rgba(255, 255, 255, 0.25);
	color: white;
	border-radius: 10px;
	padding: 0 .4rem;
	font-size: .75rem;
	margin-inline-start: .35rem;
}

.subproject-chevron {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	padding: .4rem .3rem;
	background: var(--primary);
	color: white;
	border-radius: 4px;
	cursor: pointer;
	font-size: .65rem;
	transition: filter $transition;

	&:hover {
		filter: brightness(1.15);
	}
}

.subproject-dropdown-wrap {
	position: relative;
	display: inline-flex;

	:deep(.popup) {
		inset-inline-end: 0;
		inset-inline-start: auto;
		inset-block-start: 2rem;
	}
}

.subproject-popup {
	min-inline-size: 200px;
	padding: .75rem;
}

.subproject-item {
	padding-block: .15rem;
}

.subproject-label {
	display: inline-flex;
	align-items: center;
	gap: .4rem;
}

.subproject-color-dot {
	display: inline-block;
	inline-size: 10px;
	block-size: 10px;
	border-radius: 50%;
	flex-shrink: 0;
}

.subproject-legend {
	display: inline-flex;
	align-items: center;
	gap: .5rem;
	flex-wrap: wrap;
}

.legend-item {
	display: inline-flex;
	align-items: center;
	gap: .3rem;
	padding: .15rem .5rem;
	border-radius: $radius;
	font-size: .8rem;
	color: var(--text);
	cursor: pointer;
	transition: opacity $transition;
	user-select: none;

	&:hover {
		background: var(--grey-100);
	}

	&.is-excluded {
		opacity: .45;
		text-decoration: line-through;
	}

	&.is-parent {
		cursor: default;
		font-weight: 600;
		opacity: 1;

		&:hover {
			background: none;
		}
	}
}

.legend-dot {
	display: inline-block;
	inline-size: 10px;
	block-size: 10px;
	border-radius: 50%;
	flex-shrink: 0;
	transition: background-color $transition;
}

.legend-label {
	white-space: nowrap;
}
</style>
