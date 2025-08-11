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
			:focused-cell="focusedCell"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, computed} from 'vue'
import {onClickOutside} from '@vueuse/core'

const props = defineProps<{ 
	rows: string[]
	cellsByRow: Record<string,string[]>
}>()
const emit = defineEmits<{
  (e:'update:focused', payload:{row:string|null;cell:number|null}):void
}>()

const chartRef = ref<HTMLElement | null>(null)
const focusedRowIndex = ref<number|null>(null)
const focusedCellIndex = ref<number|null>(null)

const focusedRow = computed(()=>focusedRowIndex.value===null?null:props.rows[focusedRowIndex.value])
const focusedCell = focusedCellIndex
const cellsCount = computed(()=>props.rows.length?props.cellsByRow[props.rows[0]].length:0)

onClickOutside(chartRef, () => {
	focusedRowIndex.value = null
	focusedCellIndex.value = null
	emit('update:focused', {row:null, cell:null})
})

function onKeyDown(e: KeyboardEvent) {
	if (focusedRowIndex.value === null || focusedCellIndex.value === null) return
	
	let newRowIndex = focusedRowIndex.value
	let newCellIndex = focusedCellIndex.value
	
	if (e.key === 'ArrowRight') {
		newCellIndex = Math.min(newCellIndex + 1, cellsCount.value - 1)
	}
	if (e.key === 'ArrowLeft') {
		newCellIndex = Math.max(newCellIndex - 1, 0)
	}
	if (e.key === 'ArrowDown') {
		newRowIndex = Math.min(newRowIndex + 1, props.rows.length - 1)
	}
	if (e.key === 'ArrowUp') {
		newRowIndex = Math.max(newRowIndex - 1, 0)
	}
	
	focusedRowIndex.value = newRowIndex
	focusedCellIndex.value = newCellIndex
	emit('update:focused', {row: focusedRow.value, cell: focusedCellIndex.value})
}

function initializeFocus() {
	// Only initialize focus if not already set and there are rows
	if (focusedRowIndex.value === null && props.rows.length > 0) {
		focusedRowIndex.value = 0
		focusedCellIndex.value = 0
		emit('update:focused', {row: focusedRow.value, cell: focusedCellIndex.value})
	}
}

function setFocus(rowId: string, cellIndex: number = 0) {
	const rowIndex = props.rows.indexOf(rowId)
	if (rowIndex !== -1) {
		focusedRowIndex.value = rowIndex
		focusedCellIndex.value = Math.max(0, Math.min(cellIndex, cellsCount.value - 1))
		emit('update:focused', {row: focusedRow.value, cell: focusedCellIndex.value})
	}
}

// Expose methods for parent components
defineExpose({
	setFocus,
	initializeFocus,
})
</script>
