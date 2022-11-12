<template>
	<div class="attachments">
		<h3>
			<span class="icon is-grey">
				<icon icon="paperclip"/>
			</span>
			{{ $t('task.attachment.title') }}
		</h3>

		<input
			v-if="editEnabled"
			:disabled="loading || undefined"
			@change="uploadNewAttachment()"
			id="files"
			multiple
			ref="filesRef"
			type="file"
		/>
		<progress
			v-if="attachmentService.uploadProgress > 0"
			:value="attachmentService.uploadProgress"
			class="progress is-primary"
			max="100"
		>
			{{ attachmentService.uploadProgress }}%
		</progress>

		<div class="files" v-if="attachments.length > 0">
			<!-- FIXME: don't use a for element that wraps other links / buttons
				Instead: overlay element with button that is inside.
			-->
			<a
				class="attachment"
				v-for="a in attachments"
				:key="a.id"
				@click="viewOrDownload(a)"
			>
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
						<i18n-t keypath="task.attachment.createdBy" scope="global">
							<span v-tooltip="formatDateLong(a.created)">
								{{ formatDateSince(a.created) }}
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
							class="attachment-info-meta-button"
							@click.prevent.stop="downloadAttachment(a)"
							v-tooltip="$t('task.attachment.downloadTooltip')"
						>
							{{ $t('misc.download') }}
						</BaseButton>
						<BaseButton
							class="attachment-info-meta-button"
							@click.stop="copyUrl(a)"
							v-tooltip="$t('task.attachment.copyUrlTooltip')"
						>
							{{ $t('task.attachment.copyUrl') }}
						</BaseButton>
						<BaseButton
							v-if="editEnabled"
							class="attachment-info-meta-button"
							@click.prevent.stop="setAttachmentToDelete(a)"
							v-tooltip="$t('task.attachment.deleteTooltip')"
						>
							{{ $t('misc.delete') }}
						</BaseButton>
						<BaseButton
							v-if="editEnabled"
							class="attachment-info-meta-button"
							@click.prevent.stop="setCoverImage(task.coverImageAttachmentId === a.id ? null : a)"
						>
							{{
								task.coverImageAttachmentId === a.id
									? $t('task.attachment.unsetAsCover')
									: $t('task.attachment.setAsCover')
							}}
						</BaseButton>
					</p>
				</div>
			</a>
		</div>

		<x-button
			v-if="editEnabled"
			:disabled="loading"
			@click="filesRef?.click()"
			class="mb-4"
			icon="cloud-upload-alt"
			variant="secondary"
			:shadow="false"
		>
			{{ $t('task.attachment.upload') }}
		</x-button>

		<!-- Dropzone -->
		<div
			:class="{ hidden: !isOverDropZone }"
			class="dropzone"
			v-if="editEnabled"
		>
			<div class="drop-hint">
				<div class="icon">
					<icon icon="cloud-upload-alt"/>
				</div>
				<div class="hint">{{ $t('task.attachment.drop') }}</div>
			</div>
		</div>

		<!-- Delete modal -->
		<modal
			:enabled="attachmentToDelete !== null"
			@close="setAttachmentToDelete(null)"
			@submit="deleteAttachment()"
		>
			<template #header>
				<span>{{ $t('task.attachment.delete') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('task.attachment.deleteText1', {filename: attachmentToDelete.file.name}) }}<br/>
					<strong class="has-text-white">{{ $t('misc.cannotBeUndone') }}</strong>
				</p>
			</template>
		</modal>

		<!-- Attachment image modal -->
		<modal
			:enabled="attachmentImageBlobUrl !== null"
			@close="attachmentImageBlobUrl = null"
		>
			<img :src="attachmentImageBlobUrl" alt=""/>
		</modal>
	</div>
</template>

<script setup lang="ts">
import {ref, shallowReactive, computed} from 'vue'
import {useDropZone} from '@vueuse/core'

import User from '@/components/misc/user.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import AttachmentService from '@/services/attachment'
import {SUPPORTED_IMAGE_SUFFIX} from '@/models/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {ITask} from '@/modelTypes/ITask'

import {useAttachmentStore} from '@/stores/attachments'
import {formatDateSince, formatDateLong} from '@/helpers/time/formatDate'
import {uploadFiles, generateAttachmentUrl} from '@/helpers/attachments'
import {getHumanSize} from '@/helpers/getHumanSize'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import {error, success} from '@/message'
import {useTaskStore} from '@/stores/tasks'
import {useI18n} from 'vue-i18n'

const taskStore = useTaskStore()
const {t} = useI18n({useScope: 'global'})

const props = withDefaults(defineProps<{
	task: ITask,
	initialAttachments?: IAttachment[],
	editEnabled: boolean,
}>(), {
	editEnabled: true,
})

// FIXME: this should go through the store
const emit = defineEmits(['task-changed'])

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
	if (SUPPORTED_IMAGE_SUFFIX.some((suffix) => attachment.file.name.endsWith(suffix))) {
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
	const task = await taskStore.setCoverImage(props.task, attachment)
	emit('task-changed', task)
	success({message: t('task.attachment.successfullyChangedCoverImage')})
}
</script>

<style lang="scss" scoped>
.attachments {
	input[type=file] {
		display: none;
	}

	@media screen and (max-width: $tablet) {
		.button {
			width: 100%;
		}
	}
}

.files {
	margin-bottom: 1rem;
}

.attachment {
	margin-bottom: .5rem;
	display: block;
	transition: background-color $transition;
	border-radius: $radius;
	padding: .5rem;

	&:hover {
		background-color: var(--grey-200);
	}
}

.filename {
	font-weight: bold;
	margin-bottom: .25rem;
	color: var(--text);
}

.info {
	color: var(--grey-500);
	font-size: .9rem;

	p {
		margin-bottom: 0;
		display: flex;

		> span:not(:last-child):after,
		> button:not(:last-child):after {
			content: '·';
			padding: 0 .25rem;
		}
	}
}

.dropzone {
	position: fixed;
	background: rgba(250, 250, 250, 0.8);
	top: 0;
	left: 0;
	bottom: 0;
	right: 0;
	z-index: 100;
	text-align: center;

	&.hidden {
		display: none;
	}
}

.drop-hint {
	position: absolute;
	bottom: 0;
	left: 0;
	right: 0;

	.icon {
		width: 100%;
		font-size: 5rem;
		height: auto;
		text-shadow: var(--shadow-md);
		animation: bounce 2s infinite;

		@media (prefers-reduced-motion: reduce) {
			animation: none;
		}
	}

	.hint {
		margin: .5rem auto 2rem;
		border-radius: 2px;
		box-shadow: var(--shadow-md);
		background: var(--primary);
		padding: 1rem;
		color: var(--white);
		width: 100%;
		max-width: 300px;
	}
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
}

@keyframes bounce {
	from,
	20%,
	53%,
	80%,
	to {
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

.is-task-cover {
	background: var(--primary);
	color: var(--white);
	padding: .25rem .35rem;
	border-radius: 4px;
	font-size: .75rem;
}
</style>