<template>
	<form
		@submit.prevent="createTask"
		class="add-new-task"
	>
		<CustomTransition name="width">
			<input
				v-if="newTaskFieldActive"
				v-model="newTaskTitle"
				@blur="hideCreateNewTask"
				@keyup.esc="newTaskFieldActive = false"
				class="input"
				ref="newTaskTitleField"
				type="text"
			/>
		</CustomTransition>
		<x-button @click="showCreateTaskOrCreate" :shadow="false" icon="plus">
			{{ $t('task.new') }}
		</x-button>
	</form>
</template>

<script setup lang="ts">
import {nextTick, ref} from 'vue'
import type {ITask} from '@/modelTypes/ITask'

import CustomTransition from '@/components/misc/CustomTransition.vue'

const emit = defineEmits<{
	(e: 'create-task', title: string): Promise<ITask>
}>()

const newTaskFieldActive = ref(false)
const newTaskTitleField = ref()
const newTaskTitle = ref('')

function showCreateTaskOrCreate() {
	if (!newTaskFieldActive.value) {
		// Timeout to not send the form if the field isn't even shown
		setTimeout(() => {
			newTaskFieldActive.value = true
			nextTick(() => newTaskTitleField.value.focus())
		}, 100)
	} else {
		createTask()
	}
}

function hideCreateNewTask() {
	if (newTaskTitle.value === '') {
		nextTick(() => (newTaskFieldActive.value = false))
	}
}

async function createTask() {
	if (!newTaskFieldActive.value) {
		return
	}
	await emit('create-task', newTaskTitle.value)
	newTaskTitle.value = ''
	hideCreateNewTask()
}
</script>

<style scoped lang="scss">
.add-new-task {
	padding: 1rem .7rem .4rem .7rem;
	display: flex;
	max-width: 450px;

	.input {
		margin-right: .7rem;
		font-size: .8rem;
	}

	.button {
		font-size: .68rem;
	}
}
</style>
