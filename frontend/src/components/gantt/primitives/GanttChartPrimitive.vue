<template>
	<div
		ref="chartRef"
		role="grid"
		tabindex="0"
		:aria-rowcount="rows.length"
		:aria-colcount="cellsCount"
		@keydown="onKeyDown"
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

const props = defineProps<{ rows: string[]; cellsByRow: Record<string,string[]> }>()
const emit = defineEmits<{
  (e:'update:focused', payload:{row:string|null;cell:number|null}):void
  (e:'update:selected', payload:{start:{row:string;cell:number};end:{row:string;cell:number}}):void
}>()

const chartRef = ref<HTMLElement | null>(null)
const focusedRowIndex = ref<number|null>(null)
const focusedCellIndex = ref<number|null>(null)
const anchor = ref<{row:string;cell:number}|null>(null)

const focusedRow = computed(()=>focusedRowIndex.value===null?null:props.rows[focusedRowIndex.value])
const focusedCell = focusedCellIndex
const cellsCount = computed(()=>props.rows.length?props.cellsByRow[props.rows[0]].length:0)

onClickOutside(chartRef, () => {
	focusedRowIndex.value = null
	focusedCellIndex.value = null
	anchor.value = null
	emit('update:focused', {row:null, cell:null})
})

function onKeyDown(e: KeyboardEvent) {
	if (focusedRowIndex.value === null || focusedCellIndex.value === null) return
	if (e.key === 'ArrowRight') focusedCellIndex.value++
	if (e.key === 'ArrowLeft') focusedCellIndex.value--
	if (e.key === 'ArrowDown') focusedRowIndex.value++
	if (e.key === 'ArrowUp') focusedRowIndex.value--
	emit('update:focused', {row: focusedRow.value, cell: focusedCellIndex.value})
}
</script>
