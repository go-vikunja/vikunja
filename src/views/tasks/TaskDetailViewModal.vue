<template>
	<div class="modal-mask">
		<div @mousedown.self="close()" class="modal-container">
			<div class="scrolling-content">
				<a @click="close()" class="close">
					<icon icon="times"/>
				</a>
				<task-detail-view/>
			</div>
		</div>
	</div>
</template>

<script>
import TaskDetailView from './TaskDetailView'

export default {
	name: 'TaskDetailViewModal',
	components: {
		TaskDetailView,
	},
	data() {
		return {
			lastRoute: null,
		}
	},
	beforeRouteEnter(to, from, next) {
		next(vm => {
			vm.lastRoute = from
		})
	},
	beforeRouteLeave(to, from, next) {
		if (from.name === 'task.kanban.detail' && to.name === 'task.detail') {
			this.$router.replace({name: 'task.kanban.detail', params: to.params})
			return
		}

		next()
	},
	methods: {
		close() {
			if (this.lastRoute === null) {
				this.$router.back()
			} else {
				this.$router.push(this.lastRoute)
			}
		},
	},
}
</script>
