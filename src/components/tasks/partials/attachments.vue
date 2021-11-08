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
			variant="secondary"
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
		async deleteAttachment() {
			try {
				const r = await this.attachmentService.delete(this.attachmentToDelete)
				this.$store.commit(
					'attachments/removeById',
					this.attachmentToDelete.id,
				)
				this.$message.success(r)
			} finally{
				this.showDeleteModal = false
			}
		},
		async viewOrDownload(attachment) {
			if (
				attachment.file.name.endsWith('.jpg') ||
				attachment.file.name.endsWith('.png') ||
				attachment.file.name.endsWith('.bmp') ||
				attachment.file.name.endsWith('.gif')
			) {
				this.showImageModal = true
				this.attachmentImageBlobUrl = await this.attachmentService.getBlobUrl(attachment)
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

<style lang="scss" scoped>
.attachments {
  input[type=file] {
    display: none;
  }

  .files {
    margin-bottom: 1rem;

    .attachment {
      margin-bottom: .5rem;
      display: block;
      transition: background-color $transition;
      border-radius: $radius;
      padding: .5rem;

      &:hover {
        background-color: var(--grey-200);
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
          > a:not(:last-child):after {
            content: '·';
            padding: 0 .25rem;
          }

          @media screen and (max-width: $mobile) {
            &.collapses {
              flex-direction: column;

              > span:not(:last-child):after,
              > a:not(:last-child):after {
                display: none;
              }

              .user .username {
                display: none;
              }
            }
          }
        }
      }
    }
  }

  @media screen and (max-width: $tablet) {
    .button {
      width: 100%;
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
  }
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

@include modal-transition();
</style>