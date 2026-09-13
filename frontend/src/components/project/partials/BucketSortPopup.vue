<template>
	<span
		ref="anchor"
		class="bucket-sort-popup-anchor"
	/>
	<Teleport to="body">
		<div
			v-if="open"
			ref="panel"
			class="bucket-sort-popup"
			:style="panelStyle"
		>
			<Card :title="$t('project.kanban.bucketSortTitle')">
				<p class="sort-description has-text-grey is-size-7">
					{{ $t('project.kanban.bucketSortDescription') }}
				</p>

				<p
					v-if="rows.length === 0"
					class="help"
				>
					{{ $t('project.kanban.sortEmptyHint') }}
				</p>

				<div
					v-for="(row, index) in rows"
					:key="'bucket_sort_'+index"
					class="bucket-sort-row"
				>
					<div class="select is-fullwidth">
						<select v-model="rows[index]">
							<option
								v-for="opt in availableOptions(index)"
								:key="opt.value"
								:value="opt.value"
							>
								{{ opt.label }}
							</option>
						</select>
					</div>
					<XButton
						variant="secondary"
						icon="chevron-up"
						:disabled="index === 0"
						@click.prevent="moveRow(index, -1)"
					/>
					<XButton
						variant="secondary"
						icon="chevron-down"
						:disabled="index === rows.length - 1"
						@click.prevent="moveRow(index, 1)"
					/>
					<button
						class="is-danger"
						@click.prevent="removeRow(index)"
					>
						<Icon icon="trash-alt" />
					</button>
				</div>

				<div class="is-flex is-justify-content-end mbe-3">
					<XButton
						v-if="canAddRow"
						variant="secondary"
						icon="plus"
						@click.prevent="addRow"
					>
						{{ $t('project.kanban.addSort') }}
					</XButton>
				</div>

				<div class="actions">
					<XButton
						variant="tertiary"
						@click="cancel()"
					>
						{{ $t('misc.cancel') }}
					</XButton>
					<XButton
						variant="secondary"
						@click="applyAll()"
					>
						{{ $t('project.kanban.bucketSortApplyAll') }}
					</XButton>
					<XButton
						variant="primary"
						@click="apply()"
					>
						{{ $t('sorting.apply') }}
					</XButton>
				</div>
			</Card>
		</div>
	</Teleport>
</template>

<script setup lang="ts">
import {computed, nextTick, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {onClickOutside} from '@vueuse/core'
import {computePosition, autoPlacement, offset, shift} from '@floating-ui/dom'

import XButton from '@/components/input/Button.vue'
import Card from '@/components/misc/Card.vue'

const props = defineProps<{
	modelValue: {sortBy: string[], sortOrder: string[]}
	open: boolean
}>()

const emit = defineEmits<{
	'update:open': [open: boolean]
	'apply': [value: {sortBy: string[], sortOrder: string[]}]
	applyAll: [value: {sortBy: string[], sortOrder: string[]}]
}>()

const {t} = useI18n({useScope: 'global'})

// bucket columns are `overflow: hidden` (for their rounded corners), which would
// clip an absolutely-positioned popup wider than the column. Teleporting to
// <body> and positioning with floating-ui (same approach as Dropdown.vue) avoids that.
const anchor = ref<HTMLElement>()
const panel = ref<HTMLElement>()
const panelPosition = ref({x: 0, y: 0})

async function updatePosition() {
	// The panel only exists in the DOM once v-if="open" has been applied, which
	// happens on Vue's next render pass — wait for it before reading refs.
	await nextTick()

	if (!anchor.value || !panel.value) {
		return
	}

	const {x, y} = await computePosition(anchor.value, panel.value, {
		placement: 'bottom-end',
		strategy: 'fixed',
		middleware: [
			offset(4),
			autoPlacement({
				allowedPlacements: ['bottom-end', 'top-end', 'bottom-start', 'top-start'],
				padding: 8,
			}),
			shift({padding: 8}),
		],
	})

	panelPosition.value = {x, y}
}

const panelStyle = computed(() => ({
	position: 'fixed' as const,
	left: `${panelPosition.value.x}px`,
	top: `${panelPosition.value.y}px`,
}))

onClickOutside(panel, () => {
	if (props.open) {
		emit('update:open', false)
	}
}, {ignore: [anchor]})

// Every row is a single combined "field:order" value so it can reuse the
// same option labels as the List view's sort popup.
const SORT_OPTIONS = [
	{value: 'priority:desc', labelKey: 'sorting.options.priorityDesc'},
	{value: 'priority:asc', labelKey: 'sorting.options.priorityAsc'},
	{value: 'due_date:asc', labelKey: 'sorting.options.dueDateAsc'},
	{value: 'due_date:desc', labelKey: 'sorting.options.dueDateDesc'},
	{value: 'start_date:asc', labelKey: 'sorting.options.startDateAsc'},
	{value: 'start_date:desc', labelKey: 'sorting.options.startDateDesc'},
	{value: 'end_date:asc', labelKey: 'sorting.options.endDateAsc'},
	{value: 'end_date:desc', labelKey: 'sorting.options.endDateDesc'},
	{value: 'percent_done:desc', labelKey: 'sorting.options.percentDoneDesc'},
	{value: 'percent_done:asc', labelKey: 'sorting.options.percentDoneAsc'},
	{value: 'created:desc', labelKey: 'sorting.options.createdDesc'},
	{value: 'created:asc', labelKey: 'sorting.options.createdAsc'},
	{value: 'updated:desc', labelKey: 'sorting.options.updatedDesc'},
	{value: 'updated:asc', labelKey: 'sorting.options.updatedAsc'},
	{value: 'title:asc', labelKey: 'sorting.options.titleAsc'},
	{value: 'title:desc', labelKey: 'sorting.options.titleDesc'},
]

const DEFAULT_VALUE_FOR_FIELD: Record<string, string> = {
	priority: 'priority:desc',
	due_date: 'due_date:asc',
	start_date: 'start_date:asc',
	end_date: 'end_date:asc',
	percent_done: 'percent_done:asc',
	created: 'created:asc',
	updated: 'updated:asc',
	title: 'title:asc',
}

function fieldOf(value: string) {
	return value.split(':')[0]
}

const rows = ref<string[]>([])

function sortArraysToRows(sortBy: string[], sortOrder: string[]) {
	return sortBy
		.map((field, i) => field ? `${field}:${sortOrder[i] || 'asc'}` : '')
		.filter(value => value !== '')
}

watch(() => props.modelValue, (val) => {
	rows.value = sortArraysToRows(val?.sortBy || [], val?.sortOrder || [])
}, {immediate: true})

// Re-sync and reposition whenever the popup is (re-)opened, discarding any unsaved edits from a previous open.
watch(() => props.open, (isOpen) => {
	if (isOpen) {
		rows.value = sortArraysToRows(props.modelValue?.sortBy || [], props.modelValue?.sortOrder || [])
		updatePosition()
	}
})

function availableOptions(index: number) {
	const usedElsewhere = rows.value.filter((_, i) => i !== index).map(fieldOf)
	return SORT_OPTIONS
		.filter(o => fieldOf(o.value) === fieldOf(rows.value[index]) || !usedElsewhere.includes(fieldOf(o.value)))
		.map(o => ({value: o.value, label: t(o.labelKey)}))
}

const canAddRow = computed(() => rows.value.length < Object.keys(DEFAULT_VALUE_FOR_FIELD).length)

function addRow() {
	const used = new Set(rows.value.map(fieldOf))
	const nextField = Object.keys(DEFAULT_VALUE_FOR_FIELD).find(f => !used.has(f))
	if (!nextField) {
		return
	}

	rows.value = [...rows.value, DEFAULT_VALUE_FOR_FIELD[nextField]]
	updatePosition()
}

function removeRow(index: number) {
	rows.value = rows.value.filter((_, i) => i !== index)
	updatePosition()
}

function moveRow(index: number, delta: number) {
	const newIndex = index + delta
	if (newIndex < 0 || newIndex >= rows.value.length) {
		return
	}

	const next = [...rows.value]
	;[next[index], next[newIndex]] = [next[newIndex], next[index]]
	rows.value = next
}

function rowsToSortArrays() {
	const sortBy: string[] = []
	const sortOrder: string[] = []
	for (const row of rows.value) {
		const [field, order] = row.split(':')
		sortBy.push(field)
		sortOrder.push(order)
	}
	return {sortBy, sortOrder}
}

function cancel() {
	emit('update:open', false)
}

function apply() {
	emit('apply', rowsToSortArrays())
	emit('update:open', false)
}

function applyAll() {
	emit('applyAll', rowsToSortArrays())
	emit('update:open', false)
}
</script>

<style scoped lang="scss">
.bucket-sort-popup-anchor {
	display: inline-block;
}

.bucket-sort-popup {
	z-index: 100;
	margin: 0;
	min-inline-size: 20rem;

	:deep(.card-content .content) {
		display: flex;
		flex-direction: column;
	}

	.sort-description {
		margin-block-end: 1rem;
	}

	.bucket-sort-row {
		display: flex;
		align-items: center;
		gap: .5rem;
		margin-block-end: .5rem;

		.select {
			flex: 1 1 auto;

			select {
				inline-size: 100%;
			}
		}

		> button {
			background: transparent;
			border: none;
			color: var(--danger);
			cursor: pointer;
		}
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: .5rem;
		flex-wrap: wrap;
	}
}
</style>
