<template>
	<div 
		ref="chartRef"
		role="grid"
		tabindex="0"
		:aria-rowcount="rows.length"
		:aria-colcount="cellsCount"
		@keydown="onKeyDown"
		@click="initializeFocus"
	>
		<slot 
			:focused-row="focusedRow"
			:focused-cell="focusedCellIndex" 
		/>
	</div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onClickOutside } from '@vueuse/core'

const props = defineProps<{
	rows: string[]
	cellsByRow: Record<string, string[]>
}>()
const emit = defineEmits<{
	(e: 'update:focused', payload: { row: string | null; cell: number | null }): void
	(e: 'enterPressed', payload: { row: string; cell: number }): void
}>()

const chartRef = ref<HTMLElement | null>(null)
const focusedRowIndex = ref<number | null>(null)
const focusedCellIndex = ref<number | null>(null)

const focusedRow = computed(() => focusedRowIndex.value === null
	? null
	: props.rows[focusedRowIndex.value])
const cellsCount = computed(() => props.rows.length 
	? props.cellsByRow[props.rows[0]].length 
	: 0)

onClickOutside(chartRef, () => {
	focusedRowIndex.value = null
	focusedCellIndex.value = null
	emit('update:focused', { row: null, cell: null })
})

function onKeyDown(e: KeyboardEvent) {
	if (focusedRowIndex.value === null || focusedCellIndex.value === null) return

	if (e.key === 'Enter') {
		e.preventDefault()
		emit('enterPressed', { row: focusedRow.value!, cell: focusedCellIndex.value })
		return
	}
}

function initializeFocus() {
	// Only initialize focus if not already set and there are rows
	if (focusedRowIndex.value === null && props.rows.length > 0) {
		focusedRowIndex.value = 0
		focusedCellIndex.value = 0
		emit('update:focused', { row: focusedRow.value, cell: focusedCellIndex.value })
	}
}

function setFocus(rowId: string, cellIndex: number = 0) {
	const rowIndex = props.rows.indexOf(rowId)
	if (rowIndex !== -1) {
		focusedRowIndex.value = rowIndex
		focusedCellIndex.value = Math.max(0, Math.min(cellIndex, cellsCount.value - 1))
		emit('update:focused', { row: focusedRow.value, cell: focusedCellIndex.value })
	}
}

// Expose methods for parent components
defineExpose({
	setFocus,
	initializeFocus,
})
</script>
