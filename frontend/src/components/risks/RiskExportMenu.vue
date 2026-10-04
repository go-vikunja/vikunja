<template>
	<Popup
		placement="bottom-start"
		:anchor="triggerEl"
	>
		<template #trigger="{toggle}">
			<XButton
				ref="trigger"
				icon="download"
				variant="secondary"
				:loading="busy"
				@click.prevent.stop="toggle()"
			>
				{{ $t('risks.export.title') }}
			</XButton>
		</template>
		<template #content="{isOpen}">
			<Card
				class="export-menu"
				:class="{'is-open': isOpen}"
			>
				<XButton
					variant="tertiary"
					:disabled="busy"
					@click="exportExcel"
				>
					{{ $t('risks.export.excel') }}
				</XButton>
				<XButton
					variant="tertiary"
					:disabled="busy"
					@click="openPrint"
				>
					{{ $t('risks.export.pdf') }}
				</XButton>
				<p class="help">
					{{ $t('risks.export.help') }}
				</p>
			</Card>
		</template>
	</Popup>
</template>

<script setup lang="ts">
import {computed, ref, type ComponentPublicInstance} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import Card from '@/components/misc/Card.vue'
import Popup from '@/components/misc/Popup.vue'
import XButton from '@/components/input/Button.vue'

import {fetchAllRisks, MAX_RISK_EXPORT} from '@/client/queries/risks'
import type {RiskFilters} from '@/client/queries/risks'
import {useProjects} from '@/composables/useProjects'
import {buildXlsx, XLSX_MIME} from '@/helpers/export/xlsx'
import {downloadBytes} from '@/helpers/export/download'
import {risksSheet, riskExportFileName, riskColumnsFor} from '@/helpers/export/riskExport'
import type {RiskColumnKey, RiskExportLabels, RiskExportLookups} from '@/helpers/export/riskExport'
import {riskFiltersToQuery, withEffectiveStatuses} from '@/helpers/riskFilterQuery'
import {error, success} from '@/message'

const props = withDefaults(defineProps<{
	// The project page exports its project; the register (null) exports the projects its filter selects.
	projectId?: number | null
	filters: RiskFilters
}>(), {
	projectId: null,
})

const {t} = useI18n({useScope: 'global'})
const router = useRouter()
const projectList = useProjects()

const trigger = ref<ComponentPublicInstance | null>(null)
const triggerEl = computed<HTMLElement | null>(() => (trigger.value?.$el as HTMLElement) ?? null)
const busy = ref(false)

// Both exports ask the list over all projects, narrowed to the project of the page where there is one.
const scopedFilters = computed<RiskFilters>(() => ({
	...withEffectiveStatuses(props.filters),
	projectIds: props.projectId !== null ? [props.projectId] : props.filters.projectIds,
}))

// A Project column is only useful where more than one project can be listed.
const multipleProjects = computed(() => (scopedFilters.value.projectIds?.length ?? 0) !== 1)

const scopeName = computed(() => {
	const ids = scopedFilters.value.projectIds ?? []
	if (ids.length === 1) {
		return projectList.projects[ids[0]]?.title ?? String(ids[0])
	}
	return t('risks.export.printAllProjects')
})

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

async function exportExcel() {
	if (busy.value) return
	busy.value = true
	try {
		const {risks, truncated} = await fetchAllRisks(null, scopedFilters.value)
		const sheet = risksSheet(risks, multipleProjects.value, lookups.value, labels.value)
		const fileName = riskExportFileName(scopeName.value, t('risks.export.fileName'), 'xlsx')
		downloadBytes(fileName, buildXlsx([sheet]), XLSX_MIME)
		success({message: truncated
			? t('risks.export.doneTruncated', {count: MAX_RISK_EXPORT})
			: t('risks.export.done', {count: risks.length, file: fileName})})
	} catch (e) {
		error(e)
	} finally {
		busy.value = false
	}
}

// The PDF is made by the browser from a print page: the user picks "Save as PDF" in the print dialog.
// The page reads the same filter from its URL and loads the risks itself.
function openPrint() {
	const href = router.resolve({
		name: 'risks.print',
		query: riskFiltersToQuery(scopedFilters.value),
	}).href
	window.open(href, '_blank', 'noopener')
}
</script>

<style lang="scss" scoped>
.export-menu {
	display: flex;
	flex-direction: column;
	align-items: stretch;
	gap: 0.25rem;
	min-inline-size: 15rem;

	.help {
		margin: 0.25rem 0 0;
		font-size: 0.8rem;
		color: var(--grey-500);
	}
}
</style>
