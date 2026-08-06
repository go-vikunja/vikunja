<template>
	<div class="attachments">
		<h2 class="task-section-title">
			<span class="icon is-grey">
				<Icon icon="paperclip" />
			</span>
			{{ $t('task.attachment.title') }}
		</h2>

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
			<div
				v-for="a in attachments"
				:key="a.id"
				class="attachment"
			>
				<div class="preview-column">
					<button
						class="preview-open"
						tabindex="-1"
						aria-hidden="true"
						@click="viewOrDownload(a)"
					>
						<FilePreview
							class="attachment-preview"
							:model-value="a"
						/>
					</button>
				</div>
				<div class="attachment-info-column">
					<button
						class="attachment-open"
						@click="viewOrDownload(a)"
					>
						<span class="filename">
							{{ a.file.name }}
							<span
								v-if="task.coverImageAttachmentId === a.id"
								class="is-task-cover"
							>
								{{ $t('task.attachment.usedAsCover') }}
							</span>
						</span>
					</button>
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
					<p class="attachment-actions">
						<BaseButton
							v-tooltip="$t('task.attachment.downloadTooltip')"
							:aria-label="$t('task.attachment.downloadTooltip')"
							class="attachment-info-meta-button"
							@click.prevent.stop="downloadAttachment(a)"
						>
							<Icon icon="download" />
						</BaseButton>
						<BaseButton
							v-tooltip="$t('task.attachment.copyUrlTooltip')"
							:aria-label="$t('task.attachment.copyUrlTooltip')"
							class="attachment-info-meta-button"
							@click.stop="copyUrl(a)"
						>
							<Icon icon="copy" />
						</BaseButton>
						<BaseButton
							v-if="editEnabled"
							v-tooltip="$t('task.attachment.deleteTooltip')"
							:aria-label="$t('task.attachment.deleteTooltip')"
							class="attachment-info-meta-button"
							@click.prevent.stop="setAttachmentToDelete(a)"
						>
							<Icon icon="trash-alt" />
						</BaseButton>
						<BaseButton
							v-if="editEnabled && canPreviewImage(a)"
							v-tooltip="task.coverImageAttachmentId === a.id
								? $t('task.attachment.unsetAsCover')
								: $t('task.attachment.setAsCover')"
							:aria-label="task.coverImageAttachmentId === a.id
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
		<Teleport :to="dropzoneTeleportTarget">
			<div
				v-if="editEnabled"
				:class="{hidden: !showDropzone}"
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

		<!-- Attachment PDF modal -->
		<Modal
			:enabled="attachmentPdfBlobUrl !== null"
			:wide="true"
			@close="attachmentPdfBlobUrl = null"
		>
			<iframe
				v-if="attachmentPdfBlobUrl"
				:src="attachmentPdfBlobUrl"
				class="pdf-preview-iframe"
			/>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {ref, shallowReactive, computed, watch, onMounted, onBeforeUnmount} from 'vue'
import {useDropZone} from '@vueuse/core'

import User from '@/components/misc/User.vue'
import ProgressBar from '@/components/misc/ProgressBar.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import AttachmentService from '@/services/attachment'
import {canPreviewImage, canPreviewPdf} from '@/models/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {ITask} from '@/modelTypes/ITask'

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

const emit = defineEmits<{
	'taskChanged': [ITask],
	'update:attachments': [IAttachment[]],
}>()

const EDITOR_SELECTOR = '.tiptap, .tiptap__editor, [contenteditable]'

function eventTargetsEditor(event: Event | null | undefined): boolean {
	if (!event) {
		return false
	}

	const target = event.target
	if (target instanceof HTMLElement && target.closest(EDITOR_SELECTOR)) {
		return true
	}

	if (typeof event.composedPath === 'function') {
		return event.composedPath().some(element =>
			element instanceof HTMLElement && element.matches(EDITOR_SELECTOR),
		)
	}

	return false
}

const taskStore = useTaskStore()
const {t} = useI18n({useScope: 'global'})

const attachmentService = shallowReactive(new AttachmentService())

const attachments = computed(() => props.task.attachments ?? [])

const loading = computed(() => attachmentService.loading || taskStore.isLoading)

const isDraggingFiles = ref(false)
const isDragOverEditor = ref(false)

function resetDragState() {
	isDraggingFiles.value = false
	isDragOverEditor.value = false
}

/**
 * Check if a drag event contains actual files (not text being dragged).
 * This prevents the file upload overlay from appearing when dragging text
 * from within the editor to outside it.
 */
function eventContainsFiles(event: Event | null | undefined): boolean {
	if (!event || !(event instanceof DragEvent)) {
		return false
	}
	return event.dataTransfer?.types.includes('Files') ?? false
}

const {isOverDropZone} = useDropZone(document, {
	onEnter(files, event) {
		if (!props.editEnabled) {
			return
		}

		// Only show dropzone if actual files are being dragged, not text
		if (!eventContainsFiles(event)) {
			return
		}

		isDraggingFiles.value = true
		isDragOverEditor.value = eventTargetsEditor(event)
	},
	onOver(files, event) {
		if (!props.editEnabled) {
			return
		}

		isDragOverEditor.value = eventTargetsEditor(event)
	},
	onLeave(files, event) {
		if (!props.editEnabled) {
			return
		}

		if (!isOverDropZone.value) {
			resetDragState()
			return
		}

		isDragOverEditor.value = eventTargetsEditor(event)
	},
	onDrop(files, event) {
		if (!props.editEnabled) {
			return
		}

		const dropOverEditor = eventTargetsEditor(event)
		resetDragState()

		// Ignore drops over editor - let TipTap handle them
		if (dropOverEditor || !files || files.length === 0) {
			return
		}

		uploadFilesToTask(files)
	},
})

const showDropzone = computed(() =>
	props.editEnabled && isDraggingFiles.value && !isDragOverEditor.value,
)

// A <dialog> opened with showModal() (e.g. the Kanban task detail) renders in
// the browser's top layer, so the full-screen dropzone overlay teleported to
// <body> would paint behind it regardless of z-index. Teleport it into the
// topmost open dialog instead, mirroring Notification.vue.
const dropzoneTeleportTarget = ref<string | HTMLElement>('body')
let dialogObserver: MutationObserver | null = null

function syncDropzoneTeleportTarget() {
	const dialogs = document.querySelectorAll<HTMLDialogElement>('dialog.modal-dialog[open]')
	dropzoneTeleportTarget.value = dialogs.item(dialogs.length - 1) ?? 'body'
}

onMounted(() => {
	syncDropzoneTeleportTarget()
	dialogObserver = new MutationObserver(syncDropzoneTeleportTarget)
	dialogObserver.observe(document.body, {
		attributes: true,
		attributeFilter: ['open'],
		childList: true,
		subtree: true,
	})
})

onBeforeUnmount(() => {
	dialogObserver?.disconnect()
	dialogObserver = null
})

watch(() => props.editEnabled, enabled => {
	if (!enabled) {
		resetDragState()
	}
})

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

async function uploadFilesToTask(files: File[] | FileList) {
	try {
		const uploaded = await uploadFiles(attachmentService, props.task.id, files)
		if (uploaded.length > 0) {
			emit('update:attachments', [...attachments.value, ...uploaded])
		}
	} catch (e) {
		error(e)
	}
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
		const updated = attachments.value.filter(a => a.id !== attachmentToDelete.value!.id)
		emit('update:attachments', updated)
		success(r)
		setAttachmentToDelete(null)
	} catch (e) {
		error(e)
	}
}

const attachmentImageBlobUrl = ref<string | null>(null)
const attachmentPdfBlobUrl = ref<string | null>(null)

async function viewOrDownload(attachment: IAttachment) {
	if (canPreviewImage(attachment)) {
		attachmentImageBlobUrl.value = await attachmentService.getBlobUrl(attachment)
	} else if (canPreviewPdf(attachment)) {
		attachmentPdfBlobUrl.value = await attachmentService.getBlobUrl(attachment)
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

defineExpose({
	openFilePicker: () => filesRef.value?.click(),
})
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

.attachment-open {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	inline-size: 100%;
	min-inline-size: 0;
	padding: 0;
	border: 0;
	background: transparent;
	color: inherit;
	font: inherit;
	text-align: start;
	cursor: pointer;
}

.filename {
	display: flex;
	align-items: center;
	font-weight: bold;
	min-block-size: 2rem;
	color: var(--text);
	text-align: start;
	word-break: break-all;
	min-inline-size: 0;
}

.attachment-info-meta,
.attachment-actions {
	color: var(--grey-500);
	font-size: .9rem;
}

.attachment-actions {
	display: flex;
	margin-block-start: .25rem;
	margin-block-end: 0;
}

.dropzone {
	position: fixed;
	background: hsla(var(--grey-100-hsl), 0.8);
	inset-block-start: 0;
	inset-inline-start: 0;
	inset-block-end: 0;
	inset-inline-end: 0;
	z-index: 4001; // above app chrome when teleported to body (no modal open)
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
	min-inline-size: 0;
}

.attachment-info-meta {
	display: flex;
	align-items: center;
	margin-block: 0;

	> span {
		padding: 0 .25rem;
	}

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

// Redundant mouse-only click target; the real control is button.attachment-open.
.preview-open {
	display: block;
	inline-size: 100%;
	block-size: 100%;
	padding: 0;
	border: 0;
	background: transparent;
	cursor: pointer;
}

.attachment-preview {
	block-size: 100%;
}

.pdf-preview-iframe {
	inline-size: 100%;
	max-inline-size: calc(100% - 4rem);
	block-size: calc(100vh - 40px);
	border: none;
	margin: 0 auto;
	display: block;
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
