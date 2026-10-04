<template>
	<div class="risk-list">
		<table class="table is-striped is-hoverable is-fullwidth">
			<thead>
				<tr>
					<th
						v-for="column in columns"
						:key="column.key"
						:aria-sort="ariaSort(column.sort)"
					>
						<button
							v-if="column.sort"
							type="button"
							class="risk-list__sort"
							@click="emit('sort', column.sort)"
						>
							{{ column.label }}
							<span
								v-if="sortBy === column.sort"
								aria-hidden="true"
							>{{ orderBy === 'desc' ? '▼' : '▲' }}</span>
						</button>
						<template v-else>
							{{ column.label }}
						</template>
					</th>
				</tr>
			</thead>
			<tbody>
				<tr
					v-for="risk in risks"
					:key="risk.id"
					class="risk-list__row"
					:class="{'is-closed': risk.status === 'closed'}"
					tabindex="0"
					@click="emit('open', risk)"
					@keydown.enter.self="emit('open', risk)"
				>
					<td>{{ risk.id }}</td>
					<td v-if="showProject">
						{{ projectTitle(risk.project_id) }}
					</td>
					<td class="risk-list__title">
						{{ risk.title }}
					</td>
					<td>{{ risk.category }}</td>
					<td>
						<RiskRatingBadge
							:rating="risk.rating"
							:score="risk.score"
						/>
					</td>
					<td>{{ $t(`risks.statuses.${risk.status}`) }}</td>
					<td>{{ ownerName(risk) }}</td>
					<td>
						{{ formatDay(risk.due_date) }}
						<span
							v-if="isOverdue(risk)"
							class="tag is-danger mis-2"
						>{{ $t('risks.overdue') }}</span>
					</td>
					<td>{{ formatDay(risk.closed_at) }}</td>
				</tr>
				<tr v-if="risks.length === 0">
					<td :colspan="columns.length">
						{{ emptyText }}
					</td>
				</tr>
			</tbody>
		</table>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import dayjs from 'dayjs'

import RiskRatingBadge from '@/components/risks/RiskRatingBadge.vue'

import type {Risk, RiskSortKey} from '@/client/queries/risks'
import {isRiskOverdue} from '@/helpers/riskRating'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

const props = withDefaults(defineProps<{
	risks: Risk[]
	// The register lists several projects, so it shows which one a risk belongs to.
	showProject?: boolean
	projectTitle?: (projectId: number) => string
	sortBy?: RiskSortKey
	orderBy?: 'asc' | 'desc'
	emptyText?: string
}>(), {
	showProject: false,
	projectTitle: () => '',
	sortBy: 'score',
	orderBy: 'desc',
	emptyText: '',
})

const emit = defineEmits<{
	open: [risk: Risk]
	sort: [key: RiskSortKey]
}>()

const {t} = useI18n({useScope: 'global'})

interface Column {
	key: string
	label: string
	sort?: RiskSortKey
}

const columns = computed<Column[]>(() => [
	{key: 'id', label: t('risks.id'), sort: 'id'},
	...(props.showProject ? [{key: 'project', label: t('risks.project')}] : []),
	{key: 'title', label: t('risks.titleField'), sort: 'title' as const},
	{key: 'category', label: t('risks.category')},
	{key: 'rating', label: t('risks.rating'), sort: 'score' as const},
	{key: 'status', label: t('risks.status'), sort: 'status' as const},
	{key: 'owner', label: t('risks.owner')},
	{key: 'due', label: t('risks.dueDate'), sort: 'due_date' as const},
	{key: 'closed', label: t('risks.closedAt')},
])

function ariaSort(sort?: RiskSortKey) {
	if (!sort || sort !== props.sortBy) return undefined
	return props.orderBy === 'desc' ? 'descending' : 'ascending'
}

function ownerName(risk: Risk): string {
	return risk.owner ? (risk.owner.name || risk.owner.username) : ''
}

function formatDay(value: string | null): string {
	const date = parseDateOrNull(value)
	return date ? dayjs(date).format('L') : ''
}

// The server decides what is overdue for the filter; this is the same rule for the mark in a row.
function isOverdue(risk: Risk): boolean {
	return isRiskOverdue(risk.status, risk.due_date)
}
</script>

<style lang="scss" scoped>
.risk-list {
	overflow-x: auto;
}

.risk-list__row {
	cursor: pointer;

	&.is-closed {
		color: var(--grey-500);
	}
}

.risk-list__title {
	font-weight: 600;
	min-inline-size: 14rem;
}

.risk-list__sort {
	background: none;
	border: 0;
	padding: 0;
	font: inherit;
	font-weight: inherit;
	color: inherit;
	cursor: pointer;
}
</style>
