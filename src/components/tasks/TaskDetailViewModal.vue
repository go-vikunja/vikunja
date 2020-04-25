<template>
	<div class="modal-mask">
		<div class="modal-container" @click.self="close()">
			<div class="scrolling-content">
				<a @click="close()" class="close">
					<icon icon="times"/>
				</a>
				<task-detail-view :parent-list="list" :parent-namespace="namespace"/>
			</div>
		</div>
	</div>
</template>

<script>
	import TaskDetailView from './TaskDetailView'
	import router from '../../router'

	export default {
		name: 'TaskDetailViewModal',
		data() {
			return {
				list: null,
				namespace: null,
			}
		},
		components: {
			TaskDetailView,
		},
		methods: {
			close() {
				router.back()
			},
		},
	}
</script>

<style scoped>

</style>