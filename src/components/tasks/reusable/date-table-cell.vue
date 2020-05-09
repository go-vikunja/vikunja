<template>
	<td v-tooltip="+date === 0 ? '' : formatDate(date)">
		{{ +date === 0 ? '-' : formatDateSince(date) }}
	</td>
</template>

<script>
	export default {
		name: 'date-table-cell',
		props: {
			date: {
				type: Date,
				default: 0,
			}
		},
	}
</script>
