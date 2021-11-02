<template>
	<div class="content has-text-centered">
		<ShowTasks
			:end-date="endDate"
			:start-date="startDate"
		/>
	</div>
</template>

<script setup>
import { ref } from 'vue'
import ShowTasks from './ShowTasks'

function getNextWeekDate() {
	return new Date((new Date()).getTime() + 7 * 24 * 60 * 60 * 1000)
}

const startDate = ref(new Date())
const endDate = ref(getNextWeekDate())
</script>