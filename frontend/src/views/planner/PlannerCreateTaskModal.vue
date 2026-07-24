<template>
	<Modal
		:enabled="true"
		@close="$emit('close')"
	>
		<Card
			:title="$t('planner.createTitle')"
			:show-close="true"
			@close="$emit('close')"
		>
			<p class="create-context">
				{{ context }}
			</p>
			<AddTask
				ref="addTaskRef"
				@taskAdded="task => $emit('created', task)"
			/>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue'

import type {ITask} from '@/modelTypes/ITask'
import Modal from '@/components/misc/Modal.vue'
import Card from '@/components/misc/Card.vue'
import AddTask from '@/components/tasks/AddTask.vue'

defineProps<{
	// Human-readable date/time the new task will be scheduled at.
	context: string
}>()

defineEmits<{
	created: [task: ITask]
	close: []
}>()

// AddTask autofocuses on mount, but the modal's showModal() runs after that and
// pulls focus to the dialog. A rAF lands after that synchronous focus move, so
// the user can start typing immediately.
const addTaskRef = ref<InstanceType<typeof AddTask> | null>(null)
onMounted(() => requestAnimationFrame(() => addTaskRef.value?.focusTaskInput()))
</script>

<style lang="scss" scoped>
.create-context {
	color: var(--grey-500);
	font-size: .9rem;
	margin-block-end: .75rem;
}
</style>
