<template>
	<div class="attachments">
		<h3>
			<span class="icon is-grey">
				<icon icon="paperclip"/>
			</span>
			{{ $t('task.attachment.title') }}
		</h3>

		<input
			:disabled="attachmentService.loading || null"
			@change="uploadNewAttachment()"
			id="files"
			multiple
			ref="files"
			type="file"
			v-if="editEnabled"
		/>
		<progress
			:value="attachmentService.uploadProgress"
			class="progress is-primary"
			max="100"
			v-if="attachmentService.uploadProgress > 0"
		>
			{{ attachmentService.uploadProgress }}%
		</progress>

		<div class="files" v-if="attachments.length > 0">
			<a
				class="attachment"
				v-for="a in attachments"
				:key="a.id"
				@click="viewOrDownload(a)"
			>
				<div class="filename">{{ a.file.name }}</div>
				<div class="info">
					<p class="collapses">
						<i18n-t keypath="task.attachment.createdBy">
							<span v-tooltip="formatDate(a.created)">
								{{ formatDateSince(a.created) }}
							</span>
							<user
								:avatar-size="24"
								:user="a.createdBy"
								:is-inline="true"
							/>
						</i18n-t>
						<span>
							{{ a.file.getHumanSize() }}
						</span>
						<span v-if="a.file.mime">
							{{ a.file.mime }}
						</span>
					</p>
					<p>
						<a
							@click.prevent.stop="downloadAttachment(a)"
							v-tooltip="$t('task.attachment.downloadTooltip')"
						>
							{{ $t('misc.download') }}
						</a>
						<a
							@click.stop="copyUrl(a)"
							v-tooltip="$t('task.attachment.copyUrlTooltip')"
						>
							{{ $t('task.attachment.copyUrl') }}
						</a>
						<a
							@click.prevent.stop="() => {attachmentToDelete = a; showDeleteModal = true}"
							v-if="editEnabled"
							v-tooltip="$t('task.attachment.deleteTooltip')"
						>
							{{ $t('misc.delete') }}
						</a>
					</p>
				</div>
			</a>
		</div>

		<x-button
			v-if="editEnabled"
			:disabled="attachmentService.loading"
			@click="$refs.files.click()"
			class="mb-4"
			icon="cloud-upload-alt"
			type="secondary"
			:shadow="false"
		>
			{{ $t('task.attachment.upload') }}
		</x-button>

		<!-- Dropzone -->
		<div
			:class="{ hidden: !showDropzone }"
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
		<transition name="modal">
			<modal
				@close="showDeleteModal = false"
				v-if="showDeleteModal"
				@submit="deleteAttachment()"
			>
				<template #header><span>{{ $t('task.attachment.delete') }}</span></template>
				
				<template #text>
					<p>{{ $t('task.attachment.deleteText1', {filename: attachmentToDelete.file.name}) }}<br/>
					<strong>{{ $t('task.attachment.deleteText2') }}</strong></p>
				</template>
			</modal>
		</transition>

		<transition name="modal">
			<modal
				@close="
					() => {
						showImageModal = false
						attachmentImageBlobUrl = null
					}
				"
				v-if="showImageModal"
			>
				<img :src="attachmentImageBlobUrl" alt=""/>
			</modal>
		</transition>
	</div>
</template>

<script>
import AttachmentService from '../../../services/attachment'
import AttachmentModel from '../../../models/attachment'
import User from '../../misc/user'
import {mapState} from 'vuex'
import copy from 'copy-to-clipboard'

import { uploadFiles, generateAttachmentUrl } from '@/helpers/attachments'

export default {
	name: 'attachments',
	components: {
		User,
	},
	data() {
		return {
			attachmentService: new AttachmentService(),
			showDropzone: false,

			showDeleteModal: false,
			attachmentToDelete: AttachmentModel,

			showImageModal: false,
			attachmentImageBlobUrl: null,
		}
	},
	props: {
		taskId: {
			required: true,
			type: Number,
		},
		initialAttachments: {
			type: Array,
		},
		editEnabled: {
			default: true,
		},
	},
	computed: mapState({
		attachments: (state) => state.attachments.attachments,
	}),
	mounted() {
		document.addEventListener('dragenter', (e) => {
			e.stopPropagation()
			e.preventDefault()
			this.showDropzone = true
		})

		window.addEventListener('dragleave', (e) => {
			e.stopPropagation()
			e.preventDefault()
			this.showDropzone = false
		})

		document.addEventListener('dragover', (e) => {
			e.stopPropagation()
			e.preventDefault()
			this.showDropzone = true
		})

		document.addEventListener('drop', (e) => {
			e.stopPropagation()
			e.preventDefault()

			let files = e.dataTransfer.files
			this.uploadFiles(files)
			this.showDropzone = false
		})
	},
	methods: {
		downloadAttachment(attachment) {
			this.attachmentService.download(attachment)
		},
		uploadNewAttachment() {
			if (this.$refs.files.files.length === 0) {
				return
			}

			this.uploadFiles(this.$refs.files.files)
		},
		uploadFiles(files) {
			uploadFiles(this.attachmentService, this.taskId, files)
		},
		deleteAttachment() {
			this.attachmentService
				.delete(this.attachmentToDelete)
				.then((r) => {
					this.$store.commit(
						'attachments/removeById',
						this.attachmentToDelete.id,
					)
					this.$message.success(r)
				})
				.finally(() => {
					this.showDeleteModal = false
				})
		},
		viewOrDownload(attachment) {
			if (
				attachment.file.name.endsWith('.jpg') ||
				attachment.file.name.endsWith('.png') ||
				attachment.file.name.endsWith('.bmp') ||
				attachment.file.name.endsWith('.gif')
			) {
				this.showImageModal = true
				this.attachmentService.getBlobUrl(attachment).then((url) => {
					this.attachmentImageBlobUrl = url
				})
			} else {
				this.downloadAttachment(attachment)
			}
		},
		copyUrl(attachment) {
			copy(generateAttachmentUrl(this.taskId, attachment.id))
		},
	},
}
</script>
