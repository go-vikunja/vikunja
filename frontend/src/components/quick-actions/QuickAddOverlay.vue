<template>
	<div class="quick-add-overlay">
		<QuickActions />
	</div>
</template>

<script setup lang="ts">
import {watch, onMounted} from 'vue'

import QuickActions from '@/components/quick-actions/QuickActions.vue'
import {useBaseStore} from '@/stores/base'

const baseStore = useBaseStore()

onMounted(() => {
	baseStore.setQuickActionsActive(true)
})

// When QuickActions closes (Escape, task created, etc.), tell Electron to hide the window
watch(() => baseStore.quickActionsActive, (active) => {
	if (!active) {
		if (typeof window.quickEntry?.close === 'function') {
			window.quickEntry.close()
		}
	}
})
</script>

<style lang="scss" scoped>
.quick-add-overlay {
	position: fixed;
	inset: 0;
	display: flex;
	align-items: flex-start;
	justify-content: center;
}
</style>
