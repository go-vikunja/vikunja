<template>
	<modal
		@close="close()"
		variant="scrolling"
		class="task-detail-view-modal"
	>
				<a @click="close()" class="close">
					<icon icon="times"/>
				</a>
				<task-detail-view/>
	</modal>
</template>

<script>
import TaskDetailView from './TaskDetailView'
import {computed} from 'vue'
import {useRoute} from 'vue-router'

export function useShowModal() {
	const route = useRoute()
	const historyState = computed(() => route.fullPath && window.history.state)
	const show = computed(() => historyState.value.backgroundView)
	return show
}

export default {
	name: 'TaskDetailViewModal',
	components: {
		TaskDetailView,
	},
	methods: {
		close() {
			this.$router.back()
		},
	},
}
</script>

<style lang="scss" scoped>
.close {
	position: fixed;
	top: 5px;
	right: 26px;
	color: var(--white);
	font-size: 2rem;

	@media screen and (max-width: $desktop) {
		color: var(--dark);
	}
}
</style>

<style lang="scss">
// Close icon SVG uses currentColor, change the color to keep it visible
.dark .task-detail-view-modal .close {
	color: var(--grey-900);
}
</style>