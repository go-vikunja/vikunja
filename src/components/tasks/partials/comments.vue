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
						<strong class="has-text-white">{{ $t('misc.cannotBeUndone') }}</strong>
					</p>
				</template>
			</modal>
		</transition>
	</div>
</template>

<script setup lang="ts">
import {ref, reactive, computed, shallowReactive, watch, nextTick} from 'vue'
import {useStore} from 'vuex'
import {useI18n} from 'vue-i18n'

import Editor from '@/components/input/AsyncEditor'

import TaskCommentService from '@/services/taskComment'
import TaskCommentModel from '@/models/taskComment'
import {uploadFile} from '@/helpers/attachments'
import {success} from '@/message'

const props = defineProps({
	taskId: {
		type: Number,
		required: true,
	},
	canWrite: {
		default: true,
	},
})

const {t} = useI18n()
const store = useStore()

const comments = ref<TaskCommentModel[]>([])

const showDeleteModal = ref(false)
const commentToDelete = reactive(new TaskCommentModel())

const isCommentEdit = ref(false)
const commentEdit = reactive(new TaskCommentModel())

const newComment = reactive(new TaskCommentModel())

const saved = ref(null)
const saving = ref(null)

const userAvatar = computed(() => store.state.auth.info.getAvatarUrl(48))
const enabled = computed(() => store.state.config.taskCommentsEnabled)
const actions = computed(() => {
	if (!props.canWrite) {
		return {}
	}
	return Object.fromEntries(comments.value.map((comment) => ([
		comment.id,
		[{
			action: () => toggleDelete(comment.id),
			title: t('misc.delete'),
		}],
	])))
})

function attachmentUpload(...args) {
	return uploadFile(props.taskId, ...args)
}

const taskCommentService = shallowReactive(new TaskCommentService())
async function loadComments(taskId) {
	if (!enabled.value) {
		return
	}

	newComment.taskId = taskId
	commentEdit.taskId = taskId
	commentToDelete.taskId = taskId
	comments.value = await taskCommentService.getAll({taskId})
}

watch(
	() => props.taskId,
	loadComments,
	{immediate: true},
)

const editorActive = ref(true)
const creating = ref(false)
async function addComment() {
	if (newComment.comment === '') {
		return
	}

	// This makes the editor trigger its mounted function again which makes it forget every input
	// it currently has in its textarea. This is a counter-hack to a hack inside of vue-easymde
	// which made it impossible to detect change from the outside. Therefore the component would
	// not update if new content from the outside was made available.
	// See https://github.com/NikulinIlya/vue-easymde/issues/3
	editorActive.value = false
	nextTick(() => (editorActive.value = true))
	creating.value = true

	try {
		const comment = await taskCommentService.create(newComment)
		comments.value.push(comment)
		newComment.comment = ''
		success({message: t('task.comment.addedSuccess')})
	} finally {
		creating.value = false
	}
}

function toggleEdit(comment: TaskCommentModel) {
	isCommentEdit.value = !isCommentEdit.value
	Object.assign(commentEdit, comment)
}

function toggleDelete(commentId) {
	showDeleteModal.value = !showDeleteModal.value
	commentToDelete.id = commentId
}

async function editComment() {
	if (commentEdit.comment === '') {
		return
	}

	saving.value = commentEdit.id

	commentEdit.taskId = props.taskId
	try {
		const comment = await taskCommentService.update(commentEdit)
		for (const c in comments.value) {
			if (comments.value[c].id === commentEdit.id) {
				comments.value[c] = comment
			}
		}
		saved.value = commentEdit.id
		setTimeout(() => {
			saved.value = null
		}, 2000)
	} finally {
		isCommentEdit.value = false
		saving.value = null
	}
}

async function deleteComment(commentToDelete: TaskCommentModel) {
	try {
		await taskCommentService.delete(commentToDelete)
		const index = comments.value.findIndex(({id}) => id === commentToDelete.id)
		comments.value.splice(index, 1)
	} finally {
		showDeleteModal.value = false
	}
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