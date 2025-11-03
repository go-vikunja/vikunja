<template>
	<div class="attachments">
		<h3>
			<span class="icon is-grey">
				<Icon icon="paperclip" />
			</span>
			{{ $t('task.attachment.title') }}
		</h3>

		<input
			v-if="editEnabled"
			id="files"
			ref="filesRef"
			:disabled="loading || undefined"
			multiple
			type="file"
			@change="uploadNewAttachment()"
		>

		<ProgressBar
			v-if="attachmentService.uploadProgress > 0"
			:value="attachmentService.uploadProgress * 100"
			is-primary
		/>

		<div
			v-if="attachments.length > 0"
			class="files"
		>
			<button
				v-for="a in attachments"
				:key="a.id"
				class="attachment"
				@click="viewOrDownload(a)"
			>
				<div class="preview-column">
					<FilePreview
						class="attachment-preview"
						:model-value="a"
					/>
				</div>
				<div class="attachment-info-column">
					<div class="filename">
						{{ a.file.name }}
						<span
							v-if="task.coverImageAttachmentId === a.id"
							class="is-task-cover"
						>
							{{ $t('task.attachment.usedAsCover') }}
						</span>
					</div>
					<div class="info">
						<p class="attachment-info-meta">
							<i18n-t
								keypath="task.attachment.createdBy"
								scope="global"
							>
								<span v-tooltip="formatDateLong(a.created)">
									{{ formatDisplayDate(a.created) }}
								</span>
								<User
									:avatar-size="24"
									:user="a.createdBy"
									:is-inline="true"
								/>
							</i18n-t>
							<span>
								{{ getHumanSize(a.file.size) }}
							</span>
							<span v-if="a.file.mime">
								{{ a.file.mime }}
							</span>
						</p>
						<p>
							<BaseButton
								v-tooltip="$t('task.attachment.downloadTooltip')"
								class="attachment-info-meta-button"
								@click.prevent.stop="downloadAttachment(a)"
							>
								<Icon icon="download" />
							</BaseButton>
							<BaseButton
								v-tooltip="$t('task.attachment.copyUrlTooltip')"
								class="attachment-info-meta-button"
								@click.stop="copyUrl(a)"
							>
								<Icon icon="copy" />
							</BaseButton>
							<BaseButton
								v-if="editEnabled"
								v-tooltip="$t('task.attachment.deleteTooltip')"
								class="attachment-info-meta-button"
								@click.prevent.stop="setAttachmentToDelete(a)"
							>
								<Icon icon="trash-alt" />
							</BaseButton>
							<BaseButton
								v-if="editEnabled && canPreview(a)"
								v-tooltip="task.coverImageAttachmentId === a.id
									? $t('task.attachment.unsetAsCover')
									: $t('task.attachment.setAsCover')"
								class="attachment-info-meta-button"
								@click.prevent.stop="setCoverImage(task.coverImageAttachmentId === a.id ? null : a)"
							>
								<Icon :icon="task.coverImageAttachmentId === a.id ? 'eye-slash' : 'eye'" />
							</BaseButton>
						</p>
					</div>
				</div>
			</button>
		</div>

		<XButton
			v-if="editEnabled"
			:disabled="loading"
			class="mbe-4"
			icon="cloud-upload-alt"
			variant="secondary"
			:shadow="false"
			@click="filesRef?.click()"
		>
			{{ $t('task.attachment.upload') }}
		</XButton>

		<!-- Dropzone -->
		<Teleport to="body">
			<div
				v-if="editEnabled"
				:class="{hidden: !isOverDropZone}"
				class="dropzone"
			>
				<div class="drop-hint">
					<div class="icon">
						<Icon icon="cloud-upload-alt" />
					</div>
					<div class="hint">
						{{ $t('task.attachment.drop') }}
					</div>
				</div>
			</div>
		</Teleport>

		<!-- Delete modal -->
		<Modal
			:enabled="attachmentToDelete !== null"
			@close="setAttachmentToDelete(null)"
			@submit="deleteAttachment()"
		>
			<template #header>
				<span>{{ $t('task.attachment.delete') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('task.attachment.deleteText1', {filename: attachmentToDelete.file.name}) }}<br>
					<strong class="has-text-white">{{ $t('misc.cannotBeUndone') }}</strong>
				</p>
			</template>
		</Modal>

		<!-- Attachment image modal -->
		<Modal
			:enabled="attachmentImageBlobUrl !== null"
			@close="attachmentImageBlobUrl = null"
		>
			<img
				:src="attachmentImageBlobUrl"
				alt=""
			>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {ref, shallowReactive, computed} from 'vue'
import {useDropZone} from '@vueuse/core'

import User from '@/components/misc/User.vue'
import ProgressBar from '@/components/misc/ProgressBar.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import AttachmentService from '@/services/attachment'
import {canPreview} from '@/models/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {ITask} from '@/modelTypes/ITask'

import {useAttachmentStore} from '@/stores/attachments'
import {formatDisplayDate, formatDateLong} from '@/helpers/time/formatDate'
import {uploadFiles, generateAttachmentUrl} from '@/helpers/attachments'
import {getHumanSize} from '@/helpers/getHumanSize'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import {error, success} from '@/message'
import {useTaskStore} from '@/stores/tasks'
import {useI18n} from 'vue-i18n'
import FilePreview from '@/components/tasks/partials/FilePreview.vue'

const props = withDefaults(defineProps<{
	task: ITask,
	editEnabled?: boolean,
}>(), {
	editEnabled: true,
})

// FIXME: this should go through the store
const emit = defineEmits<{
	'taskChanged': [ITask],
}>()
const taskStore = useTaskStore()
const {t} = useI18n({useScope: 'global'})

const attachmentService = shallowReactive(new AttachmentService())

const attachmentStore = useAttachmentStore()
const attachments = computed(() => attachmentStore.attachments)

const loading = computed(() => attachmentService.loading || taskStore.isLoading)

function onDrop(files: File[] | null) {
	if (files && files.length !== 0) {
		uploadFilesToTask(files)
	}
}

const {isOverDropZone} = useDropZone(document, onDrop)

function downloadAttachment(attachment: IAttachment) {
	attachmentService.download(attachment)
}

const filesRef = ref<HTMLInputElement | null>(null)

function uploadNewAttachment() {
	const files = filesRef.value?.files

	if (!files || files.length === 0) {
		return
	}

	uploadFilesToTask(files)
}

function uploadFilesToTask(files: File[] | FileList) {
	uploadFiles(attachmentService, props.task.id, files)
}

const attachmentToDelete = ref<IAttachment | null>(null)

function setAttachmentToDelete(attachment: IAttachment | null) {
	attachmentToDelete.value = attachment
}

async function deleteAttachment() {
	if (attachmentToDelete.value === null) {
		return
	}

	try {
		const r = await attachmentService.delete(attachmentToDelete.value)
		attachmentStore.removeById(attachmentToDelete.value.id)
		success(r)
		setAttachmentToDelete(null)
	} catch (e) {
		error(e)
	}
}

const attachmentImageBlobUrl = ref<string | null>(null)

async function viewOrDownload(attachment: IAttachment) {
	if (canPreview(attachment)) {
		attachmentImageBlobUrl.value = await attachmentService.getBlobUrl(attachment)
	} else {
		downloadAttachment(attachment)
	}
}

const copy = useCopyToClipboard()

function copyUrl(attachment: IAttachment) {
	copy(generateAttachmentUrl(props.task.id, attachment.id))
}

async function setCoverImage(attachment: IAttachment | null) {
	const updatedTask = await taskStore.setCoverImage(props.task, attachment)
	emit('taskChanged', updatedTask)
	success({message: t('task.attachment.successfullyChangedCoverImage')})
}
</script>

<style lang="scss" scoped>
.attachments {
	input[type="file"] {
		display: none;
	}

	@media screen and (max-width: $tablet) {
		.button {
			inline-size: 100%;
		}
	}
}

.files {
	margin-block-end: 1rem;
}

.attachment {
	display: grid;
	grid-template-columns: 9rem 1fr;
	align-items: center;
	inline-size: 100%;
	
	padding: .5rem;
	
	transition: background-color $transition;
	background-color: transparent;
	
	border: transparent;
	border-radius: $radius;

	&:hover {
		background-color: var(--grey-200);
	}
}

.filename {
	display: flex;
	align-items: center;
	font-weight: bold;
	block-size: 2rem;
	color: var(--text);
	text-align: start;
}

.info {
	color: var(--grey-500);
	font-size: .9rem;
	display: flex;
	flex-direction: column;

	p {
		margin-block-end: 0;
		display: flex;

		> span,
		> button:not(:last-child):after {
			padding: 0 .25rem;
		}
	}
}

.dropzone {
	position: fixed;
	background: hsla(var(--grey-100-hsl), 0.8);
	inset-block-start: 0;
	inset-inline-start: 0;
	inset-block-end: 0;
	inset-inline-end: 0;
	z-index: 4001; // modal z-index is 4000
	text-align: center;

	&.hidden {
		display: none;
	}
}

.drop-hint {
	position: absolute;
	inset-block-end: 0;
	inset-inline-start: 0;
	inset-inline-end: 0;

	.icon {
		inline-size: 100%;
		font-size: 5rem;
		block-size: auto;
		text-shadow: var(--shadow-md);
		animation: bounce 2s infinite;

		@media (prefers-reduced-motion: reduce) {
			animation: none;
		}
	}

	.hint {
		margin: .5rem auto 2rem;
		border-radius: $radius;
		box-shadow: var(--shadow-md);
		background: var(--primary);
		padding: 1rem;
		color: $white; // Should always be white because of the background, regardless of the theme
		inline-size: 100%;
		max-inline-size: 300px;
	}
}

.attachment-info-column {
	display: flex;
	flex-flow: column wrap;
	align-self: start;
}

.attachment-info-meta {
	display: flex;
	align-items: center;

	:deep(.user) {
		display: flex !important;
		align-items: center;
		margin: 0 .5rem;
	}

	@media screen and (max-width: $mobile) {
		flex-direction: column;
		align-items: flex-start;

		:deep(.user) {
			margin: .5rem 0;
		}

		> span:not(:last-child):after,
		> button:not(:last-child):after {
			display: none;
		}

		.user .username {
			display: none;
		}
	}
}

.attachment-info-meta-button {
	color: var(--link);
	padding: 0 .25rem;
}

@keyframes bounce {
	0%,
	20%,
	53%,
	80%,
	100% {
		animation-timing-function: cubic-bezier(0.215, 0.61, 0.355, 1);
		transform: translate3d(0, 0, 0);
	}

	40%,
	43% {
		animation-timing-function: cubic-bezier(0.755, 0.05, 0.855, 0.06);
		transform: translate3d(0, -30px, 0);
	}

	70% {
		animation-timing-function: cubic-bezier(0.755, 0.05, 0.855, 0.06);
		transform: translate3d(0, -15px, 0);
	}

	90% {
		transform: translate3d(0, -4px, 0);
	}
}

.preview-column {
	max-inline-size: 8rem;
	block-size: 5.2rem;
}

.attachment-preview {
	block-size: 100%;
}

.is-task-cover {
	background: var(--primary);
	color: var(--white);
	margin-inline-start: .25rem;
	padding: .25rem .35rem;
	border-radius: 4px;
	font-size: .75rem;
}
</style>
