<template>
	<ShowTasks v-bind="props" />
</template>

<script setup lang="ts">
import {parseDateOrString} from '@/helpers/time/parseDateOrString'
import {getNextWeekDate} from '@/helpers/time/getNextWeekDate'

import ShowTasks from '@/components/tasks/ShowTasks.vue'

definePage({
	name: 'tasks.range',
	props: route => ({
		dateFrom: parseDateOrString(route.query.from as string, new Date()),
		dateTo: parseDateOrString(route.query.to as string, getNextWeekDate()),
		showNulls: route.query.showNulls === 'true',
		showOverdue: route.query.showOverdue === 'true',
	}),
})

const props = defineProps<{
	dateFrom?: Date | string,
	dateTo?: Date | string,
	showNulls?: boolean,
	showOverdue?: boolean,
}>()
</script>