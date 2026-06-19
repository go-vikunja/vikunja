<template>
	<time
		v-if="date"
		v-tooltip="formatDateLong(date)"
		:datetime="formatISO(date)"
	>{{ displayText }}</time>
	<span v-else-if="fallback">{{ fallback }}</span>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {formatDisplayDate, formatDateSince, formatDateLong, formatISO} from '@/helpers/time/formatDate'

const props = withDefaults(defineProps<{
	date: Date | string | null | undefined,
	mode?: 'short' | 'relative',
	fallback?: string,
}>(), {
	mode: 'short',
	fallback: undefined,
})

const displayText = computed(() => {
	if (!props.date) return ''
	return props.mode === 'relative'
		? formatDateSince(props.date)
		: formatDisplayDate(props.date)
})
</script>
