<template>
	<div>
		<h3>
			<span class="icon is-grey">
				<icon icon="align-left"/>
			</span>
			{{ $t('task.attributes.description') }}
			<CustomTransition name="fade">
				<span class="is-small is-inline-flex" v-if="loading && saving">
					<span class="loader is-inline-block mr-2"></span>
					{{ $t('misc.saving') }}
				</span>
				<span class="is-small has-text-success" v-else-if="!loading && saved">
					<icon icon="check"/>
					{{ $t('misc.saved') }}
				</span>
			</CustomTransition>
		</h3>
		<editor
			:is-edit-enabled="canWrite"
			:upload-callback="attachmentUpload"
			:upload-enabled="true"
			:placeholder="$t('task.description.placeholder')"
			:empty-text="$t('task.description.empty')"
			:show-save="true"
			edit-shortcut="e"
			v-model="task.description"
			@update:model-value="save"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref,computed, watch, type PropType} from 'vue'

import CustomTransition from '@/components/misc/CustomTransition.vue'
import Editor from '@/components/input/AsyncEditor'

import type {ITask} from '@/modelTypes/ITask'
import {useTaskStore} from '@/stores/tasks'
import TaskModel from '@/models/task'


const props = defineProps({
	modelValue: {
		type: Object as PropType<ITask>,
		required: true,
	},
	attachmentUpload: {
		required: true,
	},
	canWrite: {
		type: Boolean,
		required: true,
	},
})

const emit = defineEmits(['update:modelValue'])

const task = ref<ITask>(new TaskModel())
const saved = ref(false)

// Since loading is global state, this variable ensures we're only showing the saving icon when saving the description.
const saving = ref(false)

const taskStore = useTaskStore()
const loading = computed(() => taskStore.isLoading)

watch(
	() => props.modelValue,
	(value) => {
		task.value = value
	},
	{immediate: true},
)

async function save() {
	saving.value = true

	try {
		// FIXME: don't update state from internal.
		task.value = await taskStore.update(task.value)
		emit('update:modelValue', task.value)

		saved.value = true
		setTimeout(() => {
			saved.value = false
		}, 2000)
	} finally {
		saving.value = false
	}
}
</script>

