<template>
	<modal
		@close="close()"
		variant="scrolling"
		class="task-detail-view-modal"
	>
		<BaseButton @click="close()" class="close">
			<icon icon="times"/>
		</BaseButton>
		<slot />
	</modal>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useRouter, useRoute} from 'vue-router'

import BaseButton from '@/components/base/BaseButton.vue'

const route = useRoute()
const historyState = computed(() => route.fullPath && window.history.state)

const router = useRouter()
function close() {
	if (historyState.value) {
		router.back()
	} else {
		const backdropRoute = historyState.value?.backdropView && router.resolve(historyState.value.backdropView)
		router.push(backdropRoute)
	}
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