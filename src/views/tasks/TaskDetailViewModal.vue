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

<script setup>
import TaskDetailView from './TaskDetailView'
import router from '@/router'

function close() {
	router.back()
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