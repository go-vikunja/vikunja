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
			:upload-callback="uploadCallback"
			:placeholder="$t('task.description.placeholder')"
			:show-save="true"
			edit-shortcut="e"
			v-model="description"
			@update:model-value="saveWithDelay"
			@save="save"
			:initial-mode="isEditorContentEmpty(description) ? 'edit' : 'preview'"
		/>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'

import CustomTransition from '@/components/misc/CustomTransition.vue'
import Editor from '@/components/input/AsyncEditor'

import type {ITask} from '@/modelTypes/ITask'
import {useTaskStore} from '@/stores/tasks'
import {isEditorContentEmpty} from '@/helpers/editorContentEmpty'

type AttachmentUploadFunction = (file: File, onSuccess: (attachmentUrl: string) => void) => Promise<string>

const {
	modelValue,
	attachmentUpload,
	canWrite,
} = defineProps<{
	modelValue: ITask,
	attachmentUpload: AttachmentUploadFunction,
	canWrite: boolean,
}>()

const emit = defineEmits(['update:modelValue'])

const description = ref<string>('')
const saved = ref(false)

// Since loading is global state, this variable ensures we're only showing the saving icon when saving the description.
const saving = ref(false)

const taskStore = useTaskStore()
const loading = computed(() => taskStore.isLoading)

watch(
	() => modelValue.description,
	value => {
		description.value = value
	},
	{immediate: true},
)

const changeTimeout = ref<ReturnType<typeof setTimeout> | null>(null)

async function saveWithDelay() {
	if (changeTimeout.value !== null) {
		clearTimeout(changeTimeout.value)
	}

	changeTimeout.value = setTimeout(async () => {
		await save()
	}, 5000)
}

async function save() {
	if (changeTimeout.value !== null) {
		clearTimeout(changeTimeout.value)
	}

	saving.value = true

	try {
		// FIXME: don't update state from internal.
		const updated = await taskStore.update({
			...modelValue,
			description: description.value,
		})
		emit('update:modelValue', updated)

		saved.value = true
		setTimeout(() => {
			saved.value = false
		}, 2000)
	} finally {
		saving.value = false
	}
}

async function uploadCallback(files: File[] | FileList): (Promise<string[]>) {

	const uploadPromises: Promise<string>[] = []

	files.forEach((file: File) => {
		const promise = new Promise<string>((resolve) => {
			attachmentUpload(file, (uploadedFileUrl: string) => resolve(uploadedFileUrl))
		})

		uploadPromises.push(promise)
	})

	return await Promise.all(uploadPromises)
}
</script>

