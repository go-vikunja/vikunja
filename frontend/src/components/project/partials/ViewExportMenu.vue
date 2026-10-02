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
				class="mie-2"
				:loading="busy"
				@click.prevent.stop="toggle()"
			>
				{{ $t('export.title') }}
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
					{{ $t('export.excel') }}
				</XButton>
				<XButton
					variant="tertiary"
					:disabled="busy"
					@click="openPrint"
				>
					{{ $t('export.pdf') }}
				</XButton>
				<p class="help">
					{{ $t('export.help') }}
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

import type {TaskFilterParams} from '@/client/queries/tasks'
import {useProjects} from '@/composables/useProjects'
import {useAuthStore} from '@/stores/auth'
import {loadTasksForExport, sortParams, type SortMap} from '@/composables/useViewExport'
import {buildXlsx, XLSX_MIME, type XlsxSheet} from '@/helpers/export/xlsx'
import {downloadBytes} from '@/helpers/export/download'
import {
	ALL_EXPORT_COLUMNS,
	capTasks,
	exportFileName,
	MAX_EXPORT_TASKS,
	sortByBoard,
	tasksSheet,
	type ExportColumnKey,
	type ExportLabels,
} from '@/helpers/export/taskExport'
import {error, success} from '@/message'

const props = withDefaults(defineProps<{
	projectId: number
	viewId: number
	kind: 'table' | 'list' | 'kanban' | 'gantt'
	// The filter and search of the view.
	params: TaskFilterParams
	sortBy?: SortMap
	// Kanban: the buckets in board order, for the Bucket column and the order of the rows.
	buckets?: {id: number, title: string}[]
	// Table: only these columns.
	columns?: ExportColumnKey[]
	// Gantt builds its own sheets from what is on screen.
	buildSheets?: () => Promise<{sheets: XlsxSheet[], truncated: boolean}>
	// Extra parameters of the print page (the Gantt date range and zoom).
	printQuery?: Record<string, string>
}>(), {
	sortBy: undefined,
	buckets: undefined,
	columns: undefined,
	buildSheets: undefined,
	printQuery: () => ({}),
})

const {t} = useI18n({useScope: 'global'})
const router = useRouter()
const projectList = useProjects()
const authStore = useAuthStore()

const trigger = ref<ComponentPublicInstance | null>(null)
const triggerEl = computed<HTMLElement | null>(() => (trigger.value?.$el as HTMLElement) ?? null)
const busy = ref(false)

const projectTitle = computed(() => projectList.projects[props.projectId]?.title ?? String(props.projectId))
const viewTitle = computed(() => t(`export.views.${props.kind}`))

const labels = computed<ExportLabels>(() => ({
	sheet: t('export.sheetTasks'),
	columns: Object.fromEntries(
		ALL_EXPORT_COLUMNS.map(column => [column.key, t(`export.columns.${column.key}`)]),
	) as ExportLabels['columns'],
}))

async function exportExcel() {
	if (busy.value) return
	busy.value = true
	try {
		let sheets: XlsxSheet[]
		let truncated: boolean

		if (props.buildSheets) {
			({sheets, truncated} = await props.buildSheets())
		} else {
			const loaded = await loadTasksForExport(props.projectId, props.viewId, props.kind, {
				...props.params,
				filter_timezone: authStore.settings.timezone,
			}, props.sortBy)
			const capped = capTasks(loaded)
			truncated = capped.truncated
			const rows = props.kind === 'kanban' && props.buckets ? sortByBoard(capped.tasks, props.buckets) : capped.tasks
			const bucketTitles = props.buckets ? new Map(props.buckets.map(b => [b.id, b.title])) : null
			sheets = [tasksSheet(rows, {
				projectTitle: id => projectList.projects[id]?.title ?? '',
				bucketTitle: bucketTitles ? id => bucketTitles.get(id) ?? '' : undefined,
			}, labels.value, props.columns)]
		}

		const fileName = exportFileName(projectTitle.value, viewTitle.value, 'xlsx')
		downloadBytes(fileName, buildXlsx(sheets), XLSX_MIME)
		success({message: truncated
			? t('export.doneTruncated', {count: MAX_EXPORT_TASKS})
			: t('export.done', {file: fileName})})
	} catch (e) {
		error(e)
	} finally {
		busy.value = false
	}
}

// The PDF is made by the browser from a print page: all scripts print with system fonts, and the
// user picks "Save as PDF" in the print dialog.
function openPrint() {
	const query: Record<string, string> = {
		kind: props.kind,
		filter: props.params.filter ?? '',
		q: props.params.q ?? '',
		filter_include_nulls: String(Boolean(props.params.filter_include_nulls)),
		...props.printQuery,
	}
	const sort = sortParams(props.sortBy)
	if (sort.sort_by) {
		query.sort = JSON.stringify(Object.fromEntries(sort.sort_by.map((key, i) => [key, (sort.order_by ?? [])[i]])))
	}
	if (props.buckets) {
		query.buckets = JSON.stringify(props.buckets)
	}
	if (props.columns) {
		query.cols = props.columns.join(',')
	}
	const href = router.resolve({
		name: 'project.print',
		params: {projectId: props.projectId, viewId: props.viewId},
		query,
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
