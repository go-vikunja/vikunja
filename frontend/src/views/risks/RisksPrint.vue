<template>
	<div class="print-page">
		<header class="print-header">
			<div>
				<h1>{{ $t('risks.export.printTitle', {scope: scopeName}) }}</h1>
				<p class="print-meta">
					<span>{{ $t('risks.export.printedOn', {date: printedOn}) }}</span>
					<span v-if="filterText"> · {{ $t('risks.export.printFilter', {filter: filterText}) }}</span>
				</p>
			</div>
			<button
				type="button"
				class="button is-primary no-print"
				@click="printNow"
			>
				{{ $t('risks.export.printNow') }}
			</button>
		</header>

		<p
			v-if="loading"
			class="print-note"
		>
			{{ $t('risks.export.printLoading') }}
		</p>
		<p
			v-else-if="failed"
			class="print-note has-text-danger"
		>
			{{ failed }}
		</p>
		<template v-else>
			<p
				v-if="truncated"
				class="print-note"
			>
				{{ $t('risks.export.printTruncated', {count: MAX_RISK_EXPORT}) }}
			</p>
			<p
				v-if="table && table.rows.length === 0"
				class="print-note"
			>
				{{ $t('risks.export.printEmpty') }}
			</p>
			<table
				v-else-if="table"
				class="print-table"
			>
				<thead>
					<tr>
						<th
							v-for="header in table.headers"
							:key="header"
						>
							{{ header }}
						</th>
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="(row, index) in table.rows"
						:key="index"
						:class="{'is-closed': table.closedRows[index]}"
					>
						<td
							v-for="(cell, cellIndex) in row"
							:key="cellIndex"
							:class="cellIndex === table.ratingColumn ? `rating is-${table.ratings[index]}` : undefined"
						>
							{{ cell }}
						</td>
					</tr>
				</tbody>
			</table>
		</template>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, onMounted, ref} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import dayjs from 'dayjs'

import {fetchAllRisks, MAX_RISK_EXPORT} from '@/client/queries/risks'
import {ensureProject} from '@/client/queries/projects'
import {useProjects} from '@/composables/useProjects'
import {
	buildRiskPrintTable,
	riskColumnsFor,
} from '@/helpers/export/riskExport'
import type {
	PrintRiskTable,
	RiskColumnKey,
	RiskExportLabels,
	RiskExportLookups,
	RiskPrintFormatters,
} from '@/helpers/export/riskExport'
import {describeRiskFilters, queryToRiskFilters, withEffectiveStatuses} from '@/helpers/riskFilterQuery'
import {getErrorText} from '@/message'

defineOptions({name: 'RisksPrint'})

const route = useRoute()
const {t} = useI18n({useScope: 'global'})
const projectList = useProjects()

const loading = ref(true)
const failed = ref('')
const truncated = ref(false)
const table = ref<PrintRiskTable | null>(null)

const filters = computed(() => queryToRiskFilters(route.query))
const printedOn = computed(() => dayjs().format('L LT'))

// One chosen project makes the project column unnecessary and names the page.
const singleProject = computed(() => (filters.value.projectIds?.length === 1 ? filters.value.projectIds[0] : null))
const scopeName = computed(() => singleProject.value !== null
	? projectList.projects[singleProject.value]?.title ?? String(singleProject.value)
	: t('risks.export.printAllProjects'))

const filterText = computed(() => describeRiskFilters(filters.value, {
	status: s => t(`risks.statuses.${s}`),
	rating: r => t(`risks.ratings.${r}`),
	search: q => t('risks.export.filterSearch', {q}),
	category: c => t('risks.export.filterCategory', {category: c}),
	overdue: t('risks.overdue'),
	unowned: t('risks.noOwner'),
	owners: t('risks.ownerSelected'),
}))

const fmt: RiskPrintFormatters = {
	date: d => dayjs(d).format('L'),
	dateTime: d => dayjs(d).format('L LT'),
}

const lookups = computed<RiskExportLookups>(() => ({
	projectTitle: id => projectList.projects[id]?.title ?? '',
	statusLabel: status => t(`risks.statuses.${status}`),
	ratingLabel: rating => t(`risks.ratings.${rating}`),
}))

const labels = computed<RiskExportLabels>(() => ({
	sheet: t('risks.export.sheet'),
	columns: Object.fromEntries(
		riskColumnsFor(true).map(column => [column.key, t(`risks.export.columns.${column.key}`)]),
	) as Record<RiskColumnKey, string>,
}))

async function load() {
	// Project titles come from the projects of the user. A project the page was not opened from is fetched.
	for (const id of filters.value.projectIds ?? []) {
		await ensureProject(id)
	}
	const result = await fetchAllRisks(null, withEffectiveStatuses(filters.value))
	truncated.value = result.truncated
	table.value = buildRiskPrintTable(result.risks, singleProject.value === null, lookups.value, fmt, labels.value)
}

function printNow() {
	window.print()
}

onMounted(async () => {
	try {
		await load()
	} catch (e) {
		failed.value = getErrorText(e)
	} finally {
		loading.value = false
	}

	// Print once the data is drawn. The button stays for another go.
	if (!failed.value) {
		await nextTick()
		requestAnimationFrame(() => setTimeout(printNow, 300))
	}
})
</script>

<style lang="scss" scoped>
.print-page {
	padding: 1rem 1.5rem;
	background: #fff;
	color: #000;
	min-block-size: 100vh;
}

.print-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
	gap: 1rem;
	margin-block-end: 1rem;

	h1 {
		font-size: 1.4rem;
		margin: 0;
	}
}

.print-meta,
.print-note {
	margin: 0.25rem 0 0;
	font-size: 0.85rem;
	color: #444;
}

.print-note {
	margin-block-end: 0.75rem;
}

.print-table {
	inline-size: 100%;
	border-collapse: collapse;
	font-size: 0.8rem;

	th,
	td {
		border: 1px solid #bbb;
		padding: 3px 6px;
		text-align: start;
		vertical-align: top;
	}

	th {
		background: #eee;
	}

	// A table that runs over a page repeats its header and never cuts a row in two.
	thead {
		display: table-header-group;
	}

	tr {
		break-inside: avoid;
		page-break-inside: avoid;
	}

	tr.is-closed td {
		color: #666;
	}

	td.rating {
		font-weight: bold;
		text-align: center;

		&.is-low {
			background: #8bc34a;
		}

		&.is-medium {
			background: #ffc107;
		}

		&.is-high {
			background: #fb8c00;
			color: #fff;
		}

		&.is-critical {
			background: #e53935;
			color: #fff;
		}
	}
}

@media print {
	.no-print {
		display: none !important;
	}

	.print-page {
		padding: 0;
		min-block-size: 0;
	}

	// The colours of the ratings carry meaning, so they stay on paper.
	.print-table th,
	.print-table td.rating {
		print-color-adjust: exact;
		-webkit-print-color-adjust: exact;
	}
}
</style>

<style lang="scss">
// The page size can only be set globally. Landscape, with a narrow margin, for the table.
@media print {
	@page {
		size: landscape;
		margin: 10mm;
	}
}
</style>
