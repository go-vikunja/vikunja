<template>
	<td v-tooltip="!date || +date === 0 ? '' : formatDateLong(date)">
		<time :datetime="date ? formatISO(date) : undefined">
			{{ !date || +date === 0 ? '-' : formatDateSince(date) }}
		</time>
	</td>
</template>

<script setup lang="ts">
import {formatISO, formatDateLong, formatDateSince} from '@/helpers/time/formatDate'

withDefaults(defineProps<{
	date?: Date | null
}>(), {
	date: null,
})
</script>
