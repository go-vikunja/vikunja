<template>
	<ProjectWrapper
		class="project-venn"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id
	>
		<template #default>
			<div
				:class="{'is-loading': loading}"
				class="loader-container"
			>
				<Card
					v-if="!loading && tasks.length === 0"
					class="has-text-centered"
				>
					<p>{{ $t('project.list.empty') }}</p>
				</Card>
				<Card
					v-else-if="!loading && allLabels.length === 0"
					class="has-text-centered"
				>
					<p>{{ $t('project.venn.noLabels') }}</p>
				</Card>
				<div
					v-else-if="layout"
					class="venn-layout"
				>
					<aside
						class="venn-facets"
						:class="{'is-open': facetsOpen}"
					>
						<XButton
							variant="secondary"
							icon="filter"
							class="venn-facets-toggle"
							:aria-expanded="facetsOpen"
							@click="facetsOpen = !facetsOpen"
						>
							{{ $t('project.venn.showLabels', {shown: shownLabelCount, total: layoutLabels.length}) }}
						</XButton>
						<Card class="venn-facets-body">
							<div class="venn-facets-header">
								<h2 class="title is-6 mbe-0">
									{{ $t('project.venn.labels') }}
								</h2>
								<span>
									<BaseButton
										class="venn-facets-link"
										@click="setHiddenLabels([])"
									>
										{{ $t('project.venn.all') }}
									</BaseButton>
									·
									<BaseButton
										class="venn-facets-link"
										@click="setHiddenLabels(layoutIds)"
									>
										{{ $t('project.venn.none') }}
									</BaseButton>
								</span>
							</div>
							<input
								v-if="allLabels.length > SEARCH_THRESHOLD"
								v-model="labelSearch"
								class="input venn-facets-search"
								type="search"
								:placeholder="$t('project.venn.searchLabels')"
								:aria-label="$t('project.venn.searchLabels')"
							>
							<ul class="venn-facets-list">
								<li
									v-for="label in listedLabels"
									:key="label.id"
								>
									<FancyCheckbox
										:model-value="isInLayout(label.id!) && !hiddenLabelIds.includes(label.id!)"
										:disabled="!isInLayout(label.id!)"
										@update:modelValue="checked => toggleLabel(label.id!, checked)"
									>
										<span
											class="venn-swatch"
											:style="{backgroundColor: labelColor(label)}"
										/>
										<span class="venn-facet-title">{{ label.title }}</span>
										<span class="venn-facet-count">{{ labelCounts.get(label.id!) ?? 0 }}</span>
									</FancyCheckbox>
								</li>
							</ul>
							<p
								v-if="allLabels.length > layoutLabels.length"
								class="help"
							>
								{{ $t('project.venn.notDrawn', {max: MAX_VENN_SETS}) }}
							</p>
						</Card>
					</aside>

					<div class="venn-main">
						<div
							v-if="board"
							class="venn-columns"
						>
							<div
								v-if="boards.length > 1"
								class="select is-small"
							>
								<select
									:value="board.id"
									:aria-label="$t('project.venn.board')"
									@change="setBoard(Number(($event.target as HTMLSelectElement).value))"
								>
									<option
										v-for="b in boards"
										:key="b.id"
										:value="b.id"
									>
										{{ b.title }}
									</option>
								</select>
							</div>
							<button
								v-for="bucket in buckets"
								:key="bucket.id"
								type="button"
								class="venn-column"
								:class="{'is-active': !hiddenBucketIds.includes(bucket.id!)}"
								:aria-pressed="!hiddenBucketIds.includes(bucket.id!)"
								@click="toggleBucket(bucket.id!)"
							>
								{{ bucket.title }}
								<span class="venn-column-count">{{ bucketCounts.get(bucket.id!) ?? 0 }}</span>
							</button>
						</div>

						<Card>
							<svg
								class="venn-diagram"
								:viewBox="`0 0 ${WIDTH} ${layout.height}`"
								role="img"
								:aria-label="$t('project.venn.chartLabel')"
							>
								<template
									v-for="(label, i) in layoutLabels"
									:key="label.id"
								>
									<g v-if="visibleMask & (1 << i)">
										<circle
											:cx="layout.circles[i].x"
											:cy="layout.circles[i].y"
											:r="layout.circles[i].r"
											:fill="labelColor(label)"
											:stroke="labelColor(label)"
											class="venn-set"
										/>
										<text
											:x="layout.titles[i].x"
											:y="layout.titles[i].y"
											:text-anchor="layout.titles[i].align"
											class="venn-set-title"
										>
											{{ label.title }}
										</text>
									</g>
								</template>
								<g
									v-for="region in diagramRegions"
									:key="region.mask"
									class="venn-region"
									:class="{'is-active': activeMask === region.mask, 'is-empty': region.tasks.length === 0}"
									role="button"
									tabindex="0"
									:aria-pressed="activeMask === region.mask"
									:aria-label="$t('project.venn.regionLabel', {region: region.name, count: region.tasks.length}, region.tasks.length)"
									@click="toggleRegion(region.mask)"
									@keydown.enter.prevent="toggleRegion(region.mask)"
									@keydown.space.prevent="toggleRegion(region.mask)"
								>
									<text
										v-if="region.mask === 0"
										:x="region.anchor.x"
										:y="region.anchor.y - 26"
										class="venn-outside-caption"
									>
										{{ region.name }}
									</text>
									<circle
										:cx="region.anchor.x"
										:cy="region.anchor.y"
										:r="region.bubble"
										class="venn-count-bubble"
									/>
									<text
										:x="region.anchor.x"
										:y="region.anchor.y"
										class="venn-count"
										:style="{fontSize: `${Math.round(region.bubble * 0.9)}px`}"
									>
										{{ region.tasks.length }}
									</text>
									<text
										v-for="(title, line) in region.titles"
										:key="line"
										:x="region.anchor.x"
										:y="region.anchor.y + region.bubble + 15 + line * 15"
										class="venn-task-title"
									>
										{{ title }}
									</text>
								</g>
							</svg>
							<p
								v-if="unplacedRegions.length > 0"
								class="venn-unplaced"
							>
								{{ $t('project.venn.unplaced') }}
								<XButton
									v-for="region in unplacedRegions"
									:key="region.mask"
									variant="tertiary"
									class="venn-unplaced-region"
									:class="{'is-active': activeMask === region.mask}"
									@click="toggleRegion(region.mask)"
								>
									{{ region.name }} ({{ region.tasks.length }})
								</XButton>
							</p>
						</Card>

						<Card :padding="false">
							<div class="venn-task-list-header">
								<h2 class="title is-5 mbe-0">
									{{ activeMask === null
										? $t('project.venn.allTasks')
										: activeMask === 0
											? $t('project.venn.outsideTasks')
											: $t('project.venn.tasksIn', {region: regionName(activeMask)}) }}
								</h2>
								<XButton
									v-if="activeMask !== null"
									variant="tertiary"
									@click="activeMask = null"
								>
									{{ $t('project.venn.showAll') }}
								</XButton>
							</div>
							<ul class="venn-task-list">
								<li
									v-for="task in listedTasks"
									:key="task.id"
									:class="{'is-done': task.done}"
								>
									<RouterLink :to="{name: 'task.detail', params: {id: task.id}}">
										<span class="venn-task-identifier">{{ getTaskIdentifier(task) }}</span>
										{{ task.title }}
									</RouterLink>
									<Labels :labels="task.labels ?? []" />
								</li>
								<li
									v-if="listedTasks.length === 0"
									class="has-text-grey"
								>
									{{ $t('project.venn.regionEmpty') }}
								</li>
							</ul>
						</Card>
					</div>
				</div>
			</div>
		</template>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {useQuery} from '@tanstack/vue-query'
import {useStorage} from '@vueuse/core'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import Labels from '@/components/tasks/partials/Labels.vue'

import type {Label} from '@/client/generated'
import {allTasksQuery, type TaskResponse} from '@/client/queries/tasks'
import {bucketsQuery} from '@/client/queries/kanban'
import {useCurrentProject} from '@/composables/useCurrentProject'
import {useDelayedLoading} from '@/composables/useDelayedLoading'
import {useAuthStore} from '@/stores/auth'
import {getHexColor, getTaskIdentifier} from '@/helpers/task'
import {
	MAX_VENN_SETS,
	groupTasksByVennRegion,
	labelsInUse,
	vennMaskOf,
	visibleMaskOf,
	type VennMask,
} from '@/helpers/vennRegions'
import {layoutEuler} from '@/helpers/eulerLayout'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: number,
	viewId: number,
}>()

const {t} = useI18n({useScope: 'global'})
const auth = useAuthStore()

const WIDTH = 600
const SEARCH_THRESHOLD = 8

const tasksQuery = useQuery(computed(() => allTasksQuery({
	project: props.projectId,
	view: props.viewId,
	params: {filter_timezone: auth.settings.timezone, expand: ['buckets']},
})))
const tasks = computed(() => tasksQuery.data.value ?? [])
const loading = useDelayedLoading(tasksQuery.isFetching)

// Per-viewer preferences, kept in the browser per view like the table columns. Hidden ids
// rather than shown ones, so new labels and columns show up without touching anything.
type ViewState = {hiddenLabels: number[], hiddenBuckets: number[], board?: number}
const storedState = useStorage<Record<number, ViewState>>('vennViewState', {})
const state = computed<ViewState>(() => {
	const stored = storedState.value[props.viewId]
	return {hiddenLabels: stored?.hiddenLabels ?? [], hiddenBuckets: stored?.hiddenBuckets ?? [], board: stored?.board}
})
function updateState(patch: Partial<ViewState>) {
	storedState.value = {...storedState.value, [props.viewId]: {...state.value, ...patch}}
}

// The diagram is laid out once from every task in the view. Hiding labels or columns only
// shows or hides circles and changes the counts on top of it.
const allLabels = computed(() => labelsInUse(tasks.value))
const layoutLabels = computed(() => allLabels.value.slice(0, MAX_VENN_SETS))
const layoutIds = computed(() => layoutLabels.value.map(label => label.id!))
const layout = computed(() => {
	if (layoutIds.value.length === 0) {
		return null
	}
	const regions = groupTasksByVennRegion(tasks.value, layoutIds.value)
	const sizes = layoutIds.value.map((_, i) => [...regions]
		.filter(([mask]) => mask & (1 << i))
		.reduce((sum, [, regionTasks]) => sum + regionTasks.length, 0))
	return layoutEuler(sizes, [...regions.keys()], WIDTH)
})

const masks = computed(() => new Map(tasks.value.map(task => [task.id, vennMaskOf(task, layoutIds.value)])))

function isInLayout(id: number) {
	return layoutIds.value.includes(id)
}

// Columns come from the project's manual kanban board: those have fixed buckets to filter by.
const {currentProject} = useCurrentProject()
const boards = computed(() => (currentProject.value?.views ?? [])
	.filter(view => view.view_kind === 'kanban' && view.bucket_configuration_mode === 'manual'))
const board = computed(() => boards.value.find(b => b.id === state.value.board) ?? boards.value[0])
const bucketsResult = useQuery(computed(() => bucketsQuery(props.projectId, board.value?.id ?? 0)))
const buckets = computed(() => [...(bucketsResult.data.value ?? [])].sort((a, b) => (a.position ?? 0) - (b.position ?? 0)))

function bucketOf(task: TaskResponse) {
	return task.buckets?.find(bucket => bucket.project_view_id === board.value?.id)?.id
}

const bucketCounts = computed(() => {
	const counts = new Map<number, number>()
	for (const task of tasks.value) {
		const bucket = bucketOf(task)
		if (bucket !== undefined) {
			counts.set(bucket, (counts.get(bucket) ?? 0) + 1)
		}
	}
	return counts
})

const hiddenBucketIds = computed(() => board.value ? state.value.hiddenBuckets : [])
const shownTasks = computed(() => {
	const hidden = new Set(hiddenBucketIds.value)
	return hidden.size === 0
		? tasks.value
		: tasks.value.filter(task => {
			const bucket = bucketOf(task)
			return bucket === undefined || !hidden.has(bucket)
		})
})

function toggleBucket(id: number) {
	const hidden = hiddenBucketIds.value
	updateState({hiddenBuckets: hidden.includes(id) ? hidden.filter(b => b !== id) : [...hidden, id]})
}

function setBoard(id: number) {
	updateState({board: id, hiddenBuckets: []})
}

const hiddenLabelIds = computed(() => state.value.hiddenLabels)
const visibleMask = computed(() => visibleMaskOf(layoutIds.value, hiddenLabelIds.value))
const shownLabelCount = computed(() => layoutIds.value.filter(id => !hiddenLabelIds.value.includes(id)).length)

function toggleLabel(id: number, checked: boolean) {
	const others = hiddenLabelIds.value.filter(hidden => hidden !== id)
	setHiddenLabels(checked ? others : [...others, id])
}

function setHiddenLabels(ids: number[]) {
	updateState({hiddenLabels: ids})
}

const labelCounts = computed(() => {
	const counts = new Map<number, number>()
	for (const task of shownTasks.value) {
		for (const id of new Set((task.labels ?? []).map(label => label.id!))) {
			counts.set(id, (counts.get(id) ?? 0) + 1)
		}
	}
	return counts
})

const labelSearch = ref('')
const listedLabels = computed(() => {
	const query = labelSearch.value.trim().toLowerCase()
	return query === ''
		? allLabels.value
		: allLabels.value.filter(label => (label.title ?? '').toLowerCase().includes(query))
})

const facetsOpen = ref(false)

function labelColor(label: Label) {
	return getHexColor(label.hex_color) ?? 'var(--grey-500)'
}

function regionName(mask: VennMask) {
	if (mask === 0) {
		return t('project.venn.outside')
	}
	return layoutLabels.value
		.filter((_, i) => mask & visibleMask.value & (1 << i))
		.map(label => label.title)
		.join(' ∩ ')
}

function regionTasks(mask: VennMask) {
	return mask === 0
		? shownTasks.value.filter(task => (masks.value.get(task.id)! & visibleMask.value) === 0)
		: shownTasks.value.filter(task => masks.value.get(task.id) === mask)
}

function truncate(title: string, length: number) {
	return title.length > length ? `${title.slice(0, length - 1)}…` : title
}

const diagramRegions = computed(() => {
	if (!layout.value) {
		return []
	}
	const regions = [...layout.value.anchors]
		.filter(([mask]) => mask & visibleMask.value)
		.map(([mask, anchor]) => {
			const margin = layout.value!.margins.get(mask) ?? 17
			const regionTaskList = regionTasks(mask)
			// Small regions get a smaller count bubble and no task titles.
			const maxTitles = margin >= 60 ? 2 : margin >= 42 ? 1 : 0
			const titles = regionTaskList.slice(0, maxTitles).map(task => truncate(task.title ?? '', 16))
			if (maxTitles > 0 && regionTaskList.length > maxTitles) {
				titles.push(t('project.venn.more', {count: regionTaskList.length - maxTitles}))
			}
			return {mask, anchor, bubble: Math.max(9, Math.min(17, margin - 3)), name: regionName(mask), tasks: regionTaskList, titles}
		})
	const outside = regionTasks(0)
	if (outside.length > 0) {
		regions.push({mask: 0, anchor: layout.value.outside, bubble: 17, name: regionName(0), tasks: outside, titles: []})
	}
	return regions
})

const unplacedRegions = computed(() => (layout.value?.unplaced ?? [])
	.filter(mask => mask & visibleMask.value)
	.map(mask => ({mask, name: regionName(mask), tasks: regionTasks(mask)})))

const activeMask = ref<VennMask | null>(null)
watch(visibleMask, mask => {
	if (activeMask.value !== null && activeMask.value !== 0 && (activeMask.value & mask) === 0) {
		activeMask.value = null
	}
})

function toggleRegion(mask: VennMask) {
	activeMask.value = activeMask.value === mask ? null : mask
}

const listedTasks = computed(() => activeMask.value === null ? shownTasks.value : regionTasks(activeMask.value))
</script>

<style lang="scss" scoped>
.venn-layout {
	display: flex;
	align-items: flex-start;
	gap: 1rem;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
		align-items: stretch;
	}
}

.venn-facets {
	flex: 0 0 15rem;
	position: sticky;
	inset-block-start: 1rem;

	@media screen and (max-width: $tablet) {
		position: static;
		flex-basis: auto;
	}
}

.venn-facets-toggle {
	display: none;
	margin-block-end: .5rem;

	@media screen and (max-width: $tablet) {
		display: inline-flex;
	}
}

.venn-facets-body {
	@media screen and (max-width: $tablet) {
		display: none;

		.venn-facets.is-open & {
			display: block;
		}
	}
}

.venn-facets-header {
	display: flex;
	align-items: baseline;
	justify-content: space-between;
	margin-block-end: .5rem;
	font-size: .875rem;
}

.venn-facets-link {
	color: var(--primary);
}

.venn-facets-search {
	margin-block-end: .5rem;
}

.venn-facets-list {
	list-style: none;
	margin: 0;
	padding: 0;
}

.venn-facets-list li {
	padding-block: .125rem;

	:deep(.fancy-checkbox) {
		display: flex;
		inline-size: 100%;
	}

	:deep(.fancy-checkbox__content) {
		display: flex;
		flex: 1;
		align-items: center;
	}
}

.venn-facet-title {
	flex: 1;
}

.venn-facet-count {
	color: var(--grey-500);
	font-size: .875rem;
	margin-inline-start: .5rem;
}

.venn-main {
	flex: 1;
	min-inline-size: 0;
}

.venn-columns {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .5rem;
	margin-block-end: 1rem;
}

.venn-column {
	display: inline-flex;
	align-items: center;
	gap: .5rem;
	padding: .25rem .75rem;
	border: 1px solid var(--grey-300);
	border-radius: 999px;
	background: var(--white);
	color: var(--grey-500);
	cursor: pointer;
	text-decoration: line-through;

	&.is-active {
		border-color: var(--primary);
		color: var(--text);
		text-decoration: none;
	}

	&:focus-visible {
		outline: 2px solid var(--primary);
		outline-offset: 2px;
	}
}

.venn-column-count {
	color: var(--grey-500);
	font-size: .875rem;
}

.venn-swatch {
	display: inline-block;
	flex: none;
	inline-size: .75rem;
	block-size: .75rem;
	border-radius: 50%;
	margin-inline-end: .375rem;
	vertical-align: middle;
}

.venn-diagram {
	display: block;
	inline-size: 100%;
	max-inline-size: 720px;
	margin: 0 auto;
}

.venn-set {
	fill-opacity: .18;
	stroke-width: 2;
}

.venn-set-title {
	fill: var(--text);
	font-weight: bold;
}

.venn-region {
	cursor: pointer;

	&:focus-visible {
		outline: none;

		.venn-count-bubble {
			stroke: var(--primary);
			stroke-width: 3;
		}
	}

	&.is-empty {
		opacity: .45;
	}

	&.is-active .venn-count-bubble {
		fill: var(--primary);
	}

	&.is-active .venn-count {
		fill: var(--white);
	}
}

.venn-count-bubble {
	fill: var(--white);
	stroke: var(--grey-400);
}

.venn-count {
	fill: var(--text);
	font-weight: bold;
	text-anchor: middle;
	dominant-baseline: central;
}

.venn-outside-caption {
	fill: var(--grey-500);
	font-size: 11px;
	font-style: italic;
	text-anchor: middle;
}

.venn-task-title {
	fill: var(--text);
	font-size: 12px;
	text-anchor: middle;
}

.venn-unplaced {
	margin-block-start: .5rem;
	text-align: center;
	color: var(--grey-500);
	font-size: .875rem;
}

.venn-unplaced-region.is-active {
	color: var(--primary);
}

.venn-task-list-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: .75rem 1rem;
	border-block-end: 1px solid var(--grey-200);
}

.venn-task-list {
	li {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: .5rem;
		padding: .5rem 1rem;
		border-block-end: 1px solid var(--grey-100);

		&:last-child {
			border-block-end: none;
		}

		&.is-done a {
			text-decoration: line-through;
			color: var(--grey-500);
		}
	}
}

.venn-task-identifier {
	color: var(--grey-500);
	margin-inline-end: .25rem;
}
</style>
