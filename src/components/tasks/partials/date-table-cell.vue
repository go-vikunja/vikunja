<template>
	<td v-tooltip="+date === 0 ? '' : formatDate(date)">
		<time :datetime="date ? formatISO(date) : null">
			{{ +date === 0 ? '-' : formatDateSince(date) }}
		</time>
	</td>
</template>

<script setup lang="ts">
defineProps({
	date: {
		type: Date,
		default: 0,
	},
})
</script>
