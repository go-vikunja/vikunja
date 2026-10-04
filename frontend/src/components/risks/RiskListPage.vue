<template>
	<Card>
		<div class="risk-page__toolbar">
			<h2 class="risk-page__title">
				{{ heading }}
			</h2>
			<div class="risk-page__actions">
				<RiskExportMenu
					:project-id="projectId"
					:filters="filters"
				/>
				<XButton
					v-if="canAdd"
					variant="primary"
					icon="plus"
					@click="creating = true"
				>
					{{ $t('risks.add') }}
				</XButton>
			</div>
		</div>

		<RiskFilters
			:model-value="filters"
			:show-projects="projectId === null"
			:projects="projectOptions"
			:current-user-id="currentUserId"
			@update:modelValue="setFilters"
		/>

		<p v-if="query.isPending.value">
			{{ $t('misc.loading') }}
		</p>
		<p
			v-else-if="query.isError.value"
			class="has-text-danger"
		>
			{{ $t('risks.loadError') }}
		</p>
		<template v-else>
			<RiskList
				:risks="risks"
				:show-project="projectId === null"
				:project-title="projectTitle"
				:sort-by="params.sortBy"
				:order-by="params.orderBy"
				:empty-text="emptyText"
				@open="open"
				@sort="sortBy"
			/>
			<PaginationEmit
				v-if="totalPages > 1"
				:total-pages="totalPages"
				:current-page="page"
				@pageChanged="p => page = p"
			/>
		</template>
	</Card>

	<RiskFormModal
		v-if="creating"
		:project-id="projectId"
		:projects="writableProjects"
		@close="creating = false"
	/>

	<RiskDetailModal
		v-if="opened"
		:initial="opened"
		@close="opened = null"
	/>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useQuery} from '@tanstack/vue-query'

import Card from '@/components/misc/Card.vue'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'
import XButton from '@/components/input/Button.vue'
import RiskDetailModal from '@/components/risks/RiskDetailModal.vue'
import RiskExportMenu from '@/components/risks/RiskExportMenu.vue'
import RiskFilters from '@/components/risks/RiskFilters.vue'
import RiskFormModal from '@/components/risks/RiskFormModal.vue'
import RiskList from '@/components/risks/RiskList.vue'

import {RISK_PAGE_SIZE, risksQuery} from '@/client/queries/risks'
import type {Risk, RiskFilters as RiskFilterValues, RiskListParams, RiskSortKey} from '@/client/queries/risks'
import {useProjects} from '@/composables/useProjects'
import {queryToRiskFilters, riskFiltersToQuery, withEffectiveStatuses} from '@/helpers/riskFilterQuery'
import {useAuthStore} from '@/stores/auth'
import {PERMISSIONS} from '@/constants/permissions'

const props = withDefaults(defineProps<{
	// null is the register over all projects.
	projectId?: number | null
}>(), {
	projectId: null,
})

const route = useRoute()
const router = useRouter()
const {t} = useI18n({useScope: 'global'})
const projectList = useProjects()
const authStore = useAuthStore()

const currentUserId = computed(() => authStore.info?.id ?? 0)

// ---- filters live in the URL ---------------------------------------------------------------
const filters = computed<RiskFilterValues>(() => queryToRiskFilters(route.query))
const page = ref(1)

function setFilters(next: RiskFilterValues) {
	router.replace({query: riskFiltersToQuery(next)})
}

// A narrower filter must not strand the view on a page that no longer exists.
watch(filters, () => {
	page.value = 1
}, {deep: true})

// A switch to another project starts again.
watch(() => props.projectId, () => {
	page.value = 1
})

const params = computed<RiskListParams>(() => ({
	...withEffectiveStatuses(filters.value),
	// The server default, written out so the header shows the arrow.
	sortBy: filters.value.sortBy ?? 'score',
	orderBy: filters.value.sortBy ? (filters.value.orderBy ?? 'asc') : 'desc',
	page: page.value,
	perPage: RISK_PAGE_SIZE,
}))

function sortBy(key: RiskSortKey) {
	const current = params.value
	// A second click on the column turns the order around. A new column starts with the natural order:
	// worst, latest and highest first for score, due date and id, alphabetical for the rest.
	const natural: 'asc' | 'desc' = key === 'score' || key === 'id' ? 'desc' : 'asc'
	const orderBy = current.sortBy === key ? (current.orderBy === 'asc' ? 'desc' : 'asc') : natural
	setFilters({...filters.value, sortBy: key, orderBy})
}

// ---- data ----------------------------------------------------------------------------------
const query = useQuery(computed(() => risksQuery(props.projectId, params.value)))
const risks = computed(() => query.data.value?.items ?? [])
const totalPages = computed(() => query.data.value?.total_pages ?? 1)

// ---- projects ------------------------------------------------------------------------------
const projectOptions = computed(() => projectList.projectsArray
	.filter(p => p.id > 0)
	.map(p => ({id: p.id, title: p.title}))
	.sort((a, b) => a.title.localeCompare(b.title)))

// Only projects the user can write to and that are not archived can take a new risk.
const writableProjects = computed(() => projectList.projectsArray
	.filter(p => p.id > 0 && !p.is_archived && (p.max_permission ?? 0) >= PERMISSIONS.READ_WRITE)
	.map(p => ({id: p.id, title: p.title})))

const currentProject = computed(() => props.projectId === null ? undefined : projectList.projects[props.projectId])

function projectTitle(id: number): string {
	return projectList.projects[id]?.title ?? ''
}

const canAdd = computed(() => props.projectId === null
	? writableProjects.value.length > 0
	: Boolean(currentProject.value && !currentProject.value.is_archived && (currentProject.value.max_permission ?? 0) >= PERMISSIONS.READ_WRITE))

const heading = computed(() => props.projectId === null
	? t('risks.register')
	: t('risks.projectTitle', {project: currentProject.value?.title ?? ''}))

const emptyText = computed(() => props.projectId === null || filtersActive.value
	? t('risks.empty')
	: t('risks.emptyProject'))

const filtersActive = computed(() => Object.keys(riskFiltersToQuery(filters.value)).some(key => key !== 'sort' && key !== 'order'))

// ---- dialogs -------------------------------------------------------------------------------
const creating = ref(false)
const opened = ref<Risk | null>(null)

function open(risk: Risk) {
	opened.value = risk
}
</script>

<style lang="scss" scoped>
.risk-page__toolbar {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	justify-content: space-between;
	gap: .75rem;
	margin-block-end: 1rem;
}

.risk-page__title {
	font-size: 1.25rem;
	font-weight: 600;
	margin: 0;
}

.risk-page__actions {
	display: flex;
	gap: .5rem;
}
</style>
