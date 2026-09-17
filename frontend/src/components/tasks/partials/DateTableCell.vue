<template>
	<td v-tooltip="!dateIsValid(date) ? '' : formatDateLong(date)">
		<time :datetime="date ? formatISO(date) : undefined">
			{{ !dateIsValid(date) ? '-' : formatDisplayDate(date) }}
		</time>
	</td>
</template>

<script setup lang="ts">
import {dateIsValid, formatISO, formatDateLong, formatDisplayDate} from '@/helpers/time/formatDate'

withDefaults(defineProps<{
	date?: Date | string
}>(), {
	date: '',
})
</script>
