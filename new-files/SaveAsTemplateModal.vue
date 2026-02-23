<template>
	<Modal
		:enabled="enabled"
		@close="$emit('close')"
	>
		<div class="save-template-content">
			<div class="modal-header">
				{{ $t('task.template.saveAsTemplate') }}
			</div>
			<div class="content">
				<p class="has-text-grey">
					{{ $t('task.template.saveAsTemplateHint') }}
				</p>

				<div class="field">
					<label class="label">{{ $t('task.template.templateName') }}</label>
					<div class="control">
						<input
							v-model="templateTitle"
							v-focus
							class="input"
							type="text"
							:placeholder="task?.title"
							@keyup.enter="saveAsTemplate"
						>
					</div>
				</div>

				<div
					v-if="errMessage"
					class="notification is-danger is-light mt-2"
				>
					{{ errMessage }}
				</div>
			</div>
			<div class="actions">
				<XButton
					variant="tertiary"
					class="has-text-danger"
					@click="$emit('close')"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="primary"
					:shadow="false"
					:loading="loading"
					@click="saveAsTemplate"
				>
					{{ $t('task.template.save') }}
				</XButton>
			</div>
		</div>
	</Modal>
</template>

<script lang="ts" setup>
import {ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import Modal from '@/components/misc/Modal.vue'

import TaskTemplateService from '@/services/taskTemplateService'
import TaskTemplateModel from '@/models/taskTemplate'

import type {ITask} from '@/modelTypes/ITask'

import {success} from '@/message'

const props = defineProps<{
	enabled: boolean,
	task: ITask | null,
}>()

const emit = defineEmits<{
	'close': [],
	'saved': [],
}>()

const {t} = useI18n({useScope: 'global'})

const templateTitle = ref('')
const loading = ref(false)
const errMessage = ref('')

watch(() => props.enabled, (newVal) => {
	if (newVal && props.task) {
		templateTitle.value = props.task.title
		errMessage.value = ''
	}
})

async function saveAsTemplate() {
	if (!props.task) return

	const title = templateTitle.value.trim() || props.task.title
	if (!title) {
		errMessage.value = t('task.template.titleRequired')
		return
	}

	loading.value = true
	errMessage.value = ''

	try {
		const service = new TaskTemplateService()
		await service.create(
			new TaskTemplateModel({
				title,
				description: props.task.description,
				priority: props.task.priority,
				hexColor: props.task.hexColor,
				percentDone: props.task.percentDone,
				labelIds: props.task.labels?.map(l => l.id) || [],
			}),
		)

		success({message: t('task.template.saveSuccess')})
		emit('saved')
		emit('close')
	} catch (e: any) {
		errMessage.value = e?.message || t('task.template.saveError')
	} finally {
		loading.value = false
	}
}
</script>

<style lang="scss" scoped>
.save-template-content {
	text-align: start;
	padding: 0 1rem;

	.modal-header {
		font-size: 2rem;
		font-weight: 700;
		text-align: center;
		margin-block-end: 1rem;
	}

	.actions {
		margin-block-start: 1.5rem;
		text-align: center;

		.button {
			margin: 0 0.5rem;
		}
	}
}

.field .label {
	color: var(--text);
	font-weight: 600;
}
</style>
