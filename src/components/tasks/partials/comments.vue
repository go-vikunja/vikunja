<template>
	<div class="content details">
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
						:has-preview="true"
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
								:has-preview="false"
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
				@close="showDeleteModal = false"
				@submit="deleteComment()"
				v-if="showDeleteModal"
			>
				<template #header><span>{{ $t('task.comment.delete') }}</span></template>
				
				<template #text>
					<p>{{ $t('task.comment.deleteText1') }}<br/>
					<strong>{{ $t('task.comment.deleteText2') }}</strong></p>
				</template>
			</modal>
		</transition>
	</div>
</template>

<script>
import TaskCommentService from '../../../services/taskComment'
import TaskCommentModel from '../../../models/taskComment'
import LoadingComponent from '../../misc/loading'
import ErrorComponent from '../../misc/error'
import { uploadFile } from '@/helpers/attachments'

export default {
	name: 'comments',
	components: {
		editor: () => ({
			component: import('../../input/editor'),
			loading: LoadingComponent,
			error: ErrorComponent,
			timeout: 60000,
		}),
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
			commentToDelete: TaskCommentModel,

			isCommentEdit: false,
			commentEdit: TaskCommentModel,

			taskCommentService: TaskCommentService,
			newComment: TaskCommentModel,
			editorActive: true,
			actions: {},

			saved: null,
			saving: null,
			creating: false,
		}
	},
	created() {
		this.taskCommentService = new TaskCommentService()
		this.newComment = new TaskCommentModel({taskId: this.taskId})
		this.commentEdit = new TaskCommentModel({taskId: this.taskId})
		this.commentToDelete = new TaskCommentModel({taskId: this.taskId})
		this.comments = []
	},
	mounted() {
		this.loadComments()
	},
	watch: {
		taskId() {
			this.loadComments()
			this.newComment.taskId = this.taskId
			this.commentEdit.taskId = this.taskId
			this.commentToDelete.taskId = this.taskId
		},
		canWrite() {
			this.makeActions()
		},
	},
	computed: {
		userAvatar() {
			return this.$store.state.auth.info.getAvatarUrl(48)
		},
	},
	methods: {
		attachmentUpload(...args) {
			return uploadFile(this.taskId, ...args)
		},

		loadComments() {
			this.taskCommentService
				.getAll({taskId: this.taskId})
				.then(r => {
					this.$set(this, 'comments', r)
					this.makeActions()
				})
				.catch((e) => {
					this.$message.error(e)
				})
		},
		addComment() {
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

			this.taskCommentService
				.create(this.newComment)
				.then((r) => {
					this.comments.push(r)
					this.newComment.comment = ''
					this.$message.success({message: this.$t('task.comment.addedSuccess')})
					this.makeActions()
				})
				.catch((e) => {
					this.$message.error(e)
				})
				.finally(() => {
					this.creating = false
				})
		},
		toggleEdit(comment) {
			this.isCommentEdit = !this.isCommentEdit
			this.commentEdit = comment
		},
		toggleDelete(commentId) {
			this.showDeleteModal = !this.showDeleteModal
			this.commentToDelete.id = commentId
		},
		editComment() {
			if (this.commentEdit.comment === '') {
				return
			}

			this.saving = this.commentEdit.id

			this.commentEdit.taskId = this.taskId
			this.taskCommentService
				.update(this.commentEdit)
				.then((r) => {
					for (const c in this.comments) {
						if (this.comments[c].id === this.commentEdit.id) {
							this.$set(this.comments, c, r)
						}
					}
					this.saved = this.commentEdit.id
					setTimeout(() => {
						this.saved = null
					}, 2000)
				})
				.catch((e) => {
					this.$message.error(e)
				})
				.finally(() => {
					this.isCommentEdit = false
					this.saving = null
				})
		},
		deleteComment() {
			this.taskCommentService
				.delete(this.commentToDelete)
				.then(() => {
					for (const a in this.comments) {
						if (this.comments[a].id === this.commentToDelete.id) {
							this.comments.splice(a, 1)
						}
					}
				})
				.catch((e) => {
					this.$message.error(e)
				})
				.finally(() => {
					this.showDeleteModal = false
				})
		},
		makeActions() {
			if (this.canWrite) {
				this.comments.forEach((c) => {
					this.$set(this.actions, c.id, [
						{
							action: () => this.toggleDelete(c.id),
							title: this.$t('misc.delete'),
						},
					])
				})
			}
		},
	},
}
</script>
