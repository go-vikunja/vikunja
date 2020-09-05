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

export default {
	name: 'ShowTasksInRange',
	components: {
		ShowTasks,
	},
	data() {
		return {
			startDate: new Date(this.$route.params.startDateUnix),
			endDate: new Date(this.$route.params.endDateUnix),
		}
	},
	watch: {
		// call again the method if the route changes
		'$route': 'setDates',
	},
	created() {
		this.setDates()
	},
	methods: {
		setDates() {
			switch (this.$route.params.type) {
				case 'week':
					this.startDate = new Date()
					this.endDate = new Date((new Date()).getTime() + 7 * 24 * 60 * 60 * 1000)
					break
				case 'month':
					this.startDate = new Date()
					this.endDate = new Date((new Date()).setMonth((new Date()).getMonth() + 1))
					break
			}
		},
	},
}
</script>