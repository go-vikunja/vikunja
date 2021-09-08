<template>
	<div class="content has-text-centered">
		<ShowTasks
			:end-date="endDate"
			:start-date="startDate"
		/>
	</div>
</template>

<script>
import ShowTasks from './ShowTasks'

function getNextWeekDate() {
	return new Date((new Date()).getTime() + 7 * 24 * 60 * 60 * 1000)
}

export default {
	name: 'ShowTasksInRange',
	components: {
		ShowTasks,
	},
	data() {
		return {
			startDate: new Date(),
			endDate: getNextWeekDate(),
		}
	},
}
</script>