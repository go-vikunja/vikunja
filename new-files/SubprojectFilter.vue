<template>
	<div
		v-if="childProjects.length > 0"
		class="subproject-filter"
	>
		<Popup>
			<template #trigger="{ toggle }">
				<XButton
					variant="secondary"
					icon="sitemap"
					:shadow="false"
					class="mis-2"
					@click.prevent.stop="toggle()"
				>
					{{ $t('task.template.subprojects') }}
					<span
						v-if="includeSubprojects"
						class="subproject-badge"
					>
						{{ enabledCount }}/{{ childProjects.length }}
					</span>
				</XButton>
			</template>
			<template #content>
				<Card class="subproject-popup">
					<div class="subproject-toggle-all">
						<FancyCheckbox
							:model-value="includeSubprojects"
							@update:modelValue="toggleAll"
						>
							{{ $t('task.template.includeSubprojects') }}
						</FancyCheckbox>
					</div>
					<div
						v-if="includeSubprojects"
						class="subproject-list"
					>
						<div
							v-for="child in childProjects"
							:key="child.id"
							class="subproject-item"
						>
							<FancyCheckbox
								:model-value="!excludedIds.has(child.id)"
								@update:modelValue="toggleProject(child.id, $event)"
							>
								{{ child.title }}
							</FancyCheckbox>
						</div>
					</div>
				</Card>
			</template>
		</Popup>
	</div>
</template>

<script lang="ts" setup>
import {ref, computed, watch, onMounted} from 'vue'

import Popup from '@/components/misc/Popup.vue'
import Card from '@/components/misc/Card.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'

import {useProjectStore} from '@/stores/projects'

import type {IProject} from '@/modelTypes/IProject'

const props = defineProps<{
	projectId: IProject['id']
}>()

const emit = defineEmits<{
	'update:includeSubprojects': [value: boolean]
	'update:excludeProjectIds': [value: string]
}>()

const projectStore = useProjectStore()

const includeSubprojects = ref(false)
const excludedIds = ref<Set<number>>(new Set())

const childProjects = computed(() => {
	if (!props.projectId || props.projectId <= 0) return []
	return Object.values(projectStore.projects)
		.filter(p => p.parentProjectId === props.projectId)
		.sort((a, b) => a.title.localeCompare(b.title))
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

function toggleProject(id: number, enabled: boolean) {
	const newSet = new Set(excludedIds.value)
	if (enabled) {
		newSet.delete(id)
	} else {
		newSet.add(id)
	}
	excludedIds.value = newSet
	emitUpdate()
}

function emitUpdate() {
	emit('update:includeSubprojects', includeSubprojects.value)
	emit('update:excludeProjectIds', Array.from(excludedIds.value).join(','))
}

// Re-emit when projectId changes
watch(() => props.projectId, () => {
	includeSubprojects.value = false
	excludedIds.value = new Set()
	emitUpdate()
})
</script>

<style lang="scss" scoped>
.subproject-filter {
	display: inline-flex;
}

.subproject-badge {
	background: var(--primary);
	color: white;
	border-radius: 10px;
	padding: 0 .4rem;
	font-size: .75rem;
	margin-inline-start: .35rem;
}

.subproject-popup {
	min-inline-size: 220px;
	padding: .75rem;
}

.subproject-toggle-all {
	padding-block-end: .5rem;
	border-block-end: 1px solid var(--grey-200);
	margin-block-end: .5rem;
	font-weight: 600;
}

.subproject-list {
	display: flex;
	flex-direction: column;
	gap: .25rem;
}

.subproject-item {
	padding-inline-start: .5rem;
}
</style>
