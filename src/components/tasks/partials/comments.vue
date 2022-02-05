<template>
	<div class="content details" v-if="enabled">
		<h3 v-if="canWrite || comments.length > 0">
			<span class="icon is-grey">
				<icon :icon="['far', 'comments']"/>
			</span>
			{{ $t('task.comment.title') }}
		</h3>
		<div class="comments">
			<span
				class="is-inline-flex is-align-items-center"
				v-if="taskCommentService.loading && saving === null && !creating"
			>
				<span class="loader is-inline-block mr-2"></span>
				{{ $t('task.comment.loading') }}
			</span>
			<div :key="c.id" class="media comment" v-for="c in comments">
				<figure class="media-left is-hidden-mobile">
					<img
						:src="c.author.getAvatarUrl(48)"
						alt=""
						class="image is-avatar"
						height="48"
						width="48"
					/>
				</figure>
				<div class="media-content">
					<div class="comment-info">
						<img
							:src="c.author.getAvatarUrl(20)"
							alt=""
							class="image is-avatar"
							height="20"
							width="20"
						/>
						<strong>{{ c.author.getDisplayName() }}</strong>&nbsp;
						<span v-tooltip="formatDate(c.created)" class="has-text-grey">
							{{ formatDateSince(c.created) }}
						</span>
						<span
							v-if="+new Date(c.created) !== +new Date(c.updated)"
							v-tooltip="formatDate(c.updated)"
						>
							· {{ $t('task.comment.edited', {date: formatDateSince(c.updated)}) }}
						</span>
						<transition name="fade">
							<span
								class="is-inline-flex"
								v-if="
									taskCommentService.loading &&
									saving === c.id
								"
							>
								<span class="loader is-inline-block mr-2"></span>
								{{ $t('misc.saving') }}
							</span>
							<span
								class="has-text-success"
								v-else-if="
									!taskCommentService.loading &&
									saved === c.id
								"
							>
								{{ $t('misc.saved') }}
							</span>
						</transition>
					</div>
					<editor
						:hasPreview="true"
						:is-edit-enabled="canWrite"
						:upload-callback="attachmentUpload"
						:upload-enabled="true"
						@change="
							() => {
								toggleEdit(c)
								editComment()
							}
						"
						v-model="c.comment"
						:bottom-actions="actions[c.id]"
						:show-save="true"
					/>
				</div>
			</div>
			<div class="media comment" v-if="canWrite">
				<figure class="media-left is-hidden-mobile">
					<img
						:src="userAvatar"
						alt=""
						class="image is-avatar"
						height="48"
						width="48"
					/>
				</figure>
				<div class="media-content">
					<div class="form">
						<transition name="fade">
							<span
								class="is-inline-flex"
								v-if="taskCommentService.loading && creating"
							>
								<span class="loader is-inline-block mr-2"></span>
								{{ $t('task.comment.creating') }}
							</span>
						</transition>
						<div class="field">
							<editor
								:class="{
									'is-loading':
										taskCommentService.loading &&
										!isCommentEdit,
								}"
								:hasPreview="false"
								:upload-callback="attachmentUpload"
								:upload-enabled="true"
								:placeholder="$t('task.comment.placeholder')"
								v-if="editorActive"
								v-model="newComment.comment"
							/>
						</div>
						<div class="field">
							<x-button
								:loading="taskCommentService.loading && !isCommentEdit"
								:disabled="newComment.comment === ''"
								@click="addComment()"
							>
								{{ $t('task.comment.comment') }}
							</x-button>
						</div>
					</div>
				</div>
			</div>
		</div>

		<transition name="modal">
			<modal
				v-if="showDeleteModal"
				@close="showDeleteModal = false"
				@submit="() => deleteComment(commentToDelete)"
			>
				<template #header><span>{{ $t('task.comment.delete') }}</span></template>

				<template #text>
					<p>
						{{ $t('task.comment.deleteText1') }}<br/>
						<strong>{{ $t('task.comment.deleteText2') }}</strong>
					</p>
				</template>
			</modal>
		</transition>
	</div>
</template>

<script>
import AsyncEditor from '@/components/input/AsyncEditor'

import TaskCommentService from '../../../services/taskComment'
import TaskCommentModel from '../../../models/taskComment'
import {uploadFile} from '@/helpers/attachments'
import {mapState} from 'vuex'

export default {
	name: 'comments',
	components: {
		Editor: AsyncEditor,
	},
	props: {
		taskId: {
			type: Number,
			required: true,
		},
		canWrite: {
			default: true,
		},
	},
	data() {
		return {
			comments: [],

			showDeleteModal: false,
			commentToDelete: new TaskCommentModel(),

			isCommentEdit: false,
			commentEdit: new TaskCommentModel(),

			taskCommentService: new TaskCommentService(),
			newComment: new TaskCommentModel(),
			editorActive: true,

			saved: null,
			saving: null,
			creating: false,
		}
	},
	watch: {
		taskId: {
			handler: 'loadComments',
			immediate: true,
		},
	},
	computed: {
		...mapState({
			userAvatar: state => state.auth.info.getAvatarUrl(48),
			enabled: state => state.config.taskCommentsEnabled,
		}),
		actions() {
			if (!this.canWrite) {
				return {}
			}
			return Object.fromEntries(this.comments.map((c) => ([
				c.id,
				[{
					action: () => this.toggleDelete(c.id),
					title: this.$t('misc.delete'),
				}],
			])))
		},
	},

	methods: {
		attachmentUpload(...args) {
			return uploadFile(this.taskId, ...args)
		},

		async loadComments(taskId) {
			if (!this.enabled) {
				return
			}

			this.newComment.taskId = taskId
			this.commentEdit.taskId = taskId
			this.commentToDelete.taskId = taskId
			this.comments = await this.taskCommentService.getAll({taskId})
		},

		async addComment() {
			if (this.newComment.comment === '') {
				return
			}

			// This makes the editor trigger its mounted function again which makes it forget every input
			// it currently has in its textarea. This is a counter-hack to a hack inside of vue-easymde
			// which made it impossible to detect change from the outside. Therefore the component would
			// not update if new content from the outside was made available.
			// See https://github.com/NikulinIlya/vue-easymde/issues/3
			this.editorActive = false
			this.$nextTick(() => (this.editorActive = true))
			this.creating = true

			try {
				const comment = await this.taskCommentService.create(this.newComment)
				this.comments.push(comment)
				this.newComment.comment = ''
				this.$message.success({message: this.$t('task.comment.addedSuccess')})
			} finally {
				this.creating = false
			}
		},

		toggleEdit(comment) {
			this.isCommentEdit = !this.isCommentEdit
			this.commentEdit = comment
		},

		toggleDelete(commentId) {
			this.showDeleteModal = !this.showDeleteModal
			this.commentToDelete.id = commentId
		},

		async editComment() {
			if (this.commentEdit.comment === '') {
				return
			}

			this.saving = this.commentEdit.id

			this.commentEdit.taskId = this.taskId
			try {
				const comment = await this.taskCommentService.update(this.commentEdit)
				for (const c in this.comments) {
					if (this.comments[c].id === this.commentEdit.id) {
						this.comments[c] = comment
					}
				}
				this.saved = this.commentEdit.id
				setTimeout(() => {
					this.saved = null
				}, 2000)
			} finally {
				this.isCommentEdit = false
				this.saving = null
			}
		},

		async deleteComment(commentToDelete) {
			try {
				await this.taskCommentService.delete(commentToDelete)
				const index = this.comments.findIndex(({id}) => id === commentToDelete.id)
				this.comments.splice(index, 1)
			} finally {
				this.showDeleteModal = false
			}
		},
	},
}
</script>

<style lang="scss" scoped>
.media-left {
	margin: 0 1rem;
}

.comment-info {
	display: flex;
	align-items: center;
	gap: .5rem;

	img {
		@media screen and (max-width: $tablet) {
			display: block;
			width: 20px;
			height: 20px;
			padding-right: 0;
			margin-right: .5rem;
		}

		@media screen and (min-width: $tablet) {
			display: none;
		}
	}


	span {
		font-size: .75rem;
		line-height: 1;
	}
}

.media-content {
	width: calc(100% - 48px - 2rem);
}

@include modal-transition();
</style>