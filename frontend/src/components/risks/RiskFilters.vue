<template>
	<div class="risk-filters">
		<div class="risk-filters__row">
			<FormInput
				v-model="searchText"
				class="risk-filters__search"
				type="search"
				:placeholder="$t('risks.search')"
				:aria-label="$t('risks.search')"
			/>
			<FormSelect
				:model-value="ownerChoice"
				:options="ownerOptions"
				:aria-label="$t('risks.owner')"
				@update:modelValue="setOwner(String($event))"
			/>
			<FormInput
				v-model="categoryText"
				type="text"
				:placeholder="$t('risks.category')"
				:aria-label="$t('risks.category')"
				@keyup.enter="applyCategory"
				@blur="applyCategory"
			/>
			<XButton
				v-if="activeCount > 0"
				variant="tertiary"
				@click="clear"
			>
				{{ $t('risks.clearFilters') }}
				<span class="tag is-info mis-2">{{ activeCount }}</span>
			</XButton>
		</div>

		<div class="risk-filters__row">
			<fieldset class="risk-filters__group">
				<legend>{{ $t('risks.status') }}</legend>
				<FormCheckbox
					v-for="status in ACTIVE_STATUSES"
					:key="status"
					:model-value="statuses.includes(status)"
					:label="$t(`risks.statuses.${status}`)"
					@update:modelValue="toggleStatus(status, $event)"
				/>
				<FormCheckbox
					:model-value="showsClosed(modelValue)"
					:label="$t('risks.showClosed')"
					@update:modelValue="update(toggleShowClosed(modelValue, $event))"
				/>
			</fieldset>

			<fieldset class="risk-filters__group">
				<legend>{{ $t('risks.rating') }}</legend>
				<FormCheckbox
					v-for="rating in RISK_RATINGS"
					:key="rating"
					:model-value="(modelValue.ratings ?? []).includes(rating)"
					:label="$t(`risks.ratings.${rating}`)"
					@update:modelValue="toggleRating(rating, $event)"
				/>
			</fieldset>

			<fieldset class="risk-filters__group">
				<legend class="is-sr-only">
					{{ $t('risks.filters') }}
				</legend>
				<FormCheckbox
					:model-value="Boolean(modelValue.overdue)"
					:label="$t('risks.onlyOverdue')"
					@update:modelValue="update({...modelValue, overdue: $event || undefined})"
				/>
				<FormCheckbox
					v-if="showProjects"
					:model-value="Boolean(modelValue.includeArchived)"
					:label="$t('risks.includeArchived')"
					@update:modelValue="update({...modelValue, includeArchived: $event || undefined})"
				/>
			</fieldset>

			<details
				v-if="showProjects && projects.length > 0"
				class="risk-filters__projects"
			>
				<summary>{{ projectSummary }}</summary>
				<div class="risk-filters__project-list">
					<FormCheckbox
						v-for="project in projects"
						:key="project.id"
						:model-value="(modelValue.projectIds ?? []).includes(project.id)"
						:label="project.title"
						@update:modelValue="toggleProject(project.id, $event)"
					/>
				</div>
			</details>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useDebounceFn} from '@vueuse/core'
import {useI18n} from 'vue-i18n'

import XButton from '@/components/input/Button.vue'
import FormCheckbox from '@/components/input/FormCheckbox.vue'
import FormInput from '@/components/input/FormInput.vue'
import FormSelect from '@/components/input/FormSelect.vue'

import type {RiskFilters} from '@/client/queries/risks'
import {
	activeRiskFilterCount,
	effectiveRiskStatuses,
	showsClosed,
	toggleShowClosed,
} from '@/helpers/riskFilterQuery'
import {RISK_RATINGS, RISK_STATUSES} from '@/helpers/riskRating'
import type {RiskRating, RiskStatus} from '@/helpers/riskRating'

const props = withDefaults(defineProps<{
	modelValue: RiskFilters
	// The register over all projects can narrow by project and include archived ones.
	showProjects?: boolean
	projects?: {id: number, title: string}[]
	currentUserId?: number
}>(), {
	showProjects: false,
	projects: () => [],
	currentUserId: 0,
})

const emit = defineEmits<{
	'update:modelValue': [value: RiskFilters]
}>()

const {t} = useI18n({useScope: 'global'})

// Closed has its own switch ("Show closed").
const ACTIVE_STATUSES = RISK_STATUSES.filter(s => s !== 'closed')

const statuses = computed(() => effectiveRiskStatuses(props.modelValue))
const activeCount = computed(() => activeRiskFilterCount(props.modelValue))

function update(next: RiskFilters) {
	emit('update:modelValue', next)
}

// ---- search --------------------------------------------------------------------------------
const searchText = ref(props.modelValue.q ?? '')
const applySearch = useDebounceFn((value: string) => {
	const q = value.trim()
	if ((props.modelValue.q ?? '') !== q) {
		update({...props.modelValue, q: q || undefined})
	}
}, 300)
watch(searchText, value => applySearch(value))
// The URL can change the filters from outside (back button, "Clear filters").
watch(() => props.modelValue.q, q => {
	if ((q ?? '') !== searchText.value.trim()) {
		searchText.value = q ?? ''
	}
})

// ---- category ------------------------------------------------------------------------------
const categoryText = ref(props.modelValue.categories?.[0] ?? '')
function applyCategory() {
	const category = categoryText.value.trim()
	if ((props.modelValue.categories?.[0] ?? '') !== category) {
		update({...props.modelValue, categories: category ? [category] : undefined})
	}
}
watch(() => props.modelValue.categories, categories => {
	categoryText.value = categories?.[0] ?? ''
})

// ---- owner ---------------------------------------------------------------------------------
const ownerChoice = computed(() => {
	const ids = props.modelValue.ownerIds ?? []
	if (ids.length === 0) return 'all'
	if (ids.length === 1 && ids[0] === 0) return 'none'
	if (ids.length === 1 && ids[0] === props.currentUserId) return 'me'
	return 'other'
})

const ownerOptions = computed(() => [
	{value: 'all', label: t('risks.owner')},
	{value: 'me', label: t('risks.ownerMe')},
	{value: 'none', label: t('risks.noOwner')},
	// A bookmarked URL can name somebody else; the option only shows then.
	...(ownerChoice.value === 'other' ? [{value: 'other', label: t('risks.ownerSelected'), disabled: true}] : []),
])

function setOwner(choice: string) {
	let ownerIds: number[] | undefined
	if (choice === 'me') ownerIds = [props.currentUserId]
	else if (choice === 'none') ownerIds = [0]
	update({...props.modelValue, ownerIds})
}

// ---- statuses, ratings, projects -----------------------------------------------------------
function toggleStatus(status: RiskStatus, on: boolean) {
	const next = on ? [...statuses.value, status] : statuses.value.filter(s => s !== status)
	const ordered = RISK_STATUSES.filter(s => next.includes(s))
	// Nothing chosen would show nothing: fall back to the default.
	update({...props.modelValue, statuses: ordered.length ? ordered : undefined})
}

function toggleRating(rating: RiskRating, on: boolean) {
	const current = props.modelValue.ratings ?? []
	const next = on ? [...current, rating] : current.filter(r => r !== rating)
	update({...props.modelValue, ratings: next.length ? RISK_RATINGS.filter(r => next.includes(r)) : undefined})
}

function toggleProject(id: number, on: boolean) {
	const current = props.modelValue.projectIds ?? []
	const next = on ? [...current, id] : current.filter(p => p !== id)
	update({...props.modelValue, projectIds: next.length ? next : undefined})
}

const projectSummary = computed(() => {
	const count = props.modelValue.projectIds?.length ?? 0
	if (count === 0) return t('risks.allProjects')
	if (count === 1) return props.projects.find(p => p.id === props.modelValue.projectIds?.[0])?.title ?? t('risks.projectsSelected', {count})
	return t('risks.projectsSelected', {count})
})

function clear() {
	// Sorting is a view setting, not a filter.
	update({sortBy: props.modelValue.sortBy, orderBy: props.modelValue.orderBy})
}
</script>

<style lang="scss" scoped>
.risk-filters {
	display: flex;
	flex-direction: column;
	gap: .75rem;
	margin-block-end: 1rem;
}

.risk-filters__row {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .75rem 1.5rem;
}

.risk-filters__search {
	min-inline-size: 14rem;
}

.risk-filters__group {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .25rem 1rem;
	border: 0;
	margin: 0;
	padding: 0;

	legend {
		float: inline-start;
		margin-inline-end: .75rem;
		font-weight: 600;
		font-size: .85rem;
	}
}

.risk-filters__projects {
	summary {
		cursor: pointer;
		font-weight: 600;
		font-size: .85rem;
	}
}

.risk-filters__project-list {
	display: flex;
	flex-direction: column;
	gap: .25rem;
	max-block-size: 14rem;
	overflow-y: auto;
	padding-block-start: .5rem;
}
</style>
