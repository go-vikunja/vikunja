<template>
	<td v-tooltip="+date === 0 ? '' : formatDateLong(date)">
		<time :datetime="date ? formatISO(date) : undefined">
			{{ +date === 0 ? '-' : formatDateSince(date) }}
		</time>
	</td>
</template>

<script setup lang="ts">
import {formatISO, formatDateLong, formatDateSince} from '@/helpers/time/formatDate'

defineProps({
	date: {
		type: Date,
		default: 0,
	},
})
</script>
