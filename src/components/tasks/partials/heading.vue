<template>
	<div class="heading">
		<BaseButton @click="copyUrl"><h1 class="title task-id">{{ textIdentifier }}</h1></BaseButton>
		<Done class="heading__done" :is-done="task.done" />
		<h1
			class="title input"
			:class="{'disabled': !canWrite}"
			@blur="save(($event.target as HTMLInputElement).textContent as string)"
			@keydown.enter.prevent.stop="($event.target as HTMLInputElement).blur()"
			:contenteditable="canWrite ? true : undefined"
			:spellcheck="false"
		>
			{{ task.title.trim() }}
		</h1>
		<transition name="fade">
			<span
				v-if="loading && saving"
				class="is-inline-flex is-align-items-center"
			>
				<span class="loader is-inline-block mr-2"></span>
				{{ $t('misc.saving') }}
			</span>
			<span
				v-else-if="!loading && showSavedMessage"
				class="has-text-success is-inline-flex is-align-content-center"
			>
				<icon icon="check" class="mr-2"/>
				{{ $t('misc.saved') }}
			</span>
		</transition>
	</div>
</template>

<script setup lang="ts">
import {ref, computed} from 'vue'
import {useStore} from 'vuex'

import BaseButton from '@/components/base/BaseButton.vue'
import Done from '@/components/misc/Done.vue'
import TaskModel from '@/models/task'
import { useRouter } from 'vue-router'
import { useCopyToClipboard } from '@/composables/useCopyToClipboard'

const props = defineProps({
	task: {
		type: TaskModel,
		required: true,
	},
	canWrite: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits(['update:task'])

const router = useRouter()
const copy = useCopyToClipboard()
async function copyUrl() {
	const route = router.resolve({ name: 'task.detail', query: { taskId: props.task.id}})
	const absoluteURL = new URL(route.href, window.location.href).href
	
	await copy(absoluteURL)
}

const store = useStore()
const loading = computed(() => store.state.loading)

const textIdentifier = computed(() => props.task?.getTextIdentifier() || '')

// Since loading is global state, this variable ensures we're only showing the saving icon when saving the description.
const saving = ref(false)

const showSavedMessage = ref(false)

async function save(title: string) {
	// We only want to save if the title was actually changed.
	// Because the contenteditable does not have a change event
	// we're building it ourselves and only continue
	// if the task title changed.
	if (title === props.task.title) {
		return
	}

	try {
		saving.value = true
		const newTask = await store.dispatch('tasks/update', {
			...props.task,
			title,
		})
		emit('update:task', newTask)
		showSavedMessage.value = true
		setTimeout(() => {
			showSavedMessage.value = false
		}, 2000)
	}
	finally {
		saving.value = false
	}
}
</script>

<style lang="scss" scoped>
.heading__done {
	margin-left: .5rem;
}
</style>