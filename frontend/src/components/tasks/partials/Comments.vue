<template>
	<div
		v-if="enabled"
		ref="commentsRef"
		class="content details comments-container"
	>
		<h3
			v-if="canWrite || comments.length > 0"
			:class="{'d-print-none': comments.length === 0}"
		>
			<span class="icon is-grey">
				<Icon :icon="['far', 'comments']" />
			</span>
			{{ $t('task.comment.title') }}
		</h3>
		<div class="comments">
			<span
				v-if="taskCommentService.loading && saving === null && !creating"
				class="is-flex is-align-items-center mbs-4 mbe-4 mis-2"
			>
				<span class="loader is-inline-block mie-2" />
				{{ $t('task.comment.loading') }}
			</span>
			<div
				v-for="c in comments"
				:id="`comment-${c.id}`"
				:key="c.id"
				class="media comment"
			>
				<figure class="media-left is-hidden-mobile">
					<img
						:src="avatarFor(c.author, 48)"
						alt=""
						class="image is-avatar"
						height="48"
						width="48"
					>
					<figcaption class="is-sr-only">
						{{ $t('misc.avatarOfUser', {user: getDisplayName(c.author)}) }}
					</figcaption>
				</figure>
				<div class="media-content">
					<div class="comment-info">
						<img
							:src="avatarFor(c.author, 20)"
							alt=""
							class="image is-avatar d-print-none"
							height="20"
							width="20"
						>
						<strong>{{ getDisplayName(c.author) }}</strong>
						<span
							v-tooltip="formatDateLong(c.created)"
							class="has-text-grey"
						>
							{{ formatDisplayDate(c.created) }}
						</span>
						<span
							v-if="+new Date(c.created) !== +new Date(c.updated)"
							v-tooltip="formatDateLong(c.updated)"
						>
							· {{ $t('task.comment.edited', {date: formatDisplayDate(c.updated)}) }}
						</span>
						<a
							v-tooltip="$t('task.comment.permalink')"
							:href="`#comment-${c.id}`"
							class="comment-permalink"
							:title="$t('task.comment.permalink')"
							@click.prevent.stop="copy(getCommentUrl(`${c.id}`))"
						>
							<span class="is-sr-only">{{ $t('task.comment.permalink') }}</span>
							<Icon icon="link" />
						</a>
						<CustomTransition name="fade">
							<span
								v-if="
									taskCommentService.loading &&
										saving === c.id
								"
								class="is-inline-flex"
							>
								<span class="loader is-inline-block mie-2" />
								{{ $t('misc.saving') }}
							</span>
							<span
								v-else-if="
									!taskCommentService.loading &&
										saved === c.id
								"
								class="has-text-success"
							>
								{{ $t('misc.saved') }}
							</span>
						</CustomTransition>
					</div>
					<Editor
						v-model="c.comment"
						:is-edit-enabled="canWrite && c.author.id === currentUserId"
						:upload-callback="attachmentUpload"
						:upload-enabled="true"
						:bottom-actions="actions[c.id]"
						:show-save="true"
						:enable-discard-shortcut="true"
						initial-mode="preview"
						@update:modelValue="
							() => {
								toggleEdit(c)
								editCommentWithDelay()
							}
						"
						@save="() => {
							toggleEdit(c)
							editComment()
						}"
					/>
					<Reactions 
						v-model="c.reactions"
						class="mbs-2" 
						entity-kind="comments"
						:entity-id="c.id"
						:disabled="!canWrite"
					/>
				</div>
			</div>

			<PaginationEmit
				v-if="taskCommentService.totalPages > 1"
				:total-pages="taskCommentService.totalPages"
				:current-page="currentPage"
				@pageChanged="changePage"
			/>

			<div
				v-if="canWrite"
				class="media comment d-print-none"
			>
				<figure class="media-left is-hidden-mobile">
					<img
						:src="userAvatar"
						alt=""
						class="image is-avatar"
						height="48"
						width="48"
					>
					<figcaption class="is-sr-only">
						{{ $t('misc.avatarOfUser', {user: getDisplayName(authStore.info)}) }}
					</figcaption>
				</figure>
				<div class="media-content">
					<div class="form">
						<CustomTransition name="fade">
							<span
								v-if="taskCommentService.loading && creating"
								class="is-inline-flex"
							>
								<span class="loader is-inline-block mie-2" />
								{{ $t('task.comment.creating') }}
							</span>
						</CustomTransition>
						<div class="field">
							<Editor
								v-if="editorActive"
								v-model="newCommentText"
								:class="{
									'is-loading':
										taskCommentService.loading &&
										!isCommentEdit,
								}"
								:upload-callback="attachmentUpload"
								:placeholder="$t('task.comment.placeholder')"
								@save="addComment()"
							/>
						</div>
						<div class="field">
							<XButton
								:loading="taskCommentService.loading && !isCommentEdit"
								:disabled="newCommentText === ''"
								@click="addComment()"
							>
								{{ $t('task.comment.comment') }}
							</XButton>
						</div>
					</div>
				</div>
			</div>
		</div>


		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="() => deleteComment(commentToDelete)"
		>
			<template #header>
				<span>{{ $t('task.comment.delete') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('task.comment.deleteText1') }}<br>
					<strong class="has-text-white">{{ $t('misc.cannotBeUndone') }}</strong>
				</p>
			</template>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {ref, reactive, computed, shallowReactive, watch, nextTick} from 'vue'
import {useI18n} from 'vue-i18n'

import CustomTransition from '@/components/misc/CustomTransition.vue'
import Editor from '@/components/input/AsyncEditor'
import PaginationEmit from '@/components/misc/PaginationEmit.vue'

import TaskCommentService from '@/services/taskComment'
import TaskCommentModel from '@/models/taskComment'

import type {ITaskComment} from '@/modelTypes/ITaskComment'
import type {ITask} from '@/modelTypes/ITask'

import {uploadFile} from '@/helpers/attachments'
import {success} from '@/message'
import {formatDateLong, formatDisplayDate} from '@/helpers/time/formatDate'
import {fetchAvatarBlobUrl, getDisplayName} from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'
import Reactions from '@/components/input/Reactions.vue'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'

const props = withDefaults(defineProps<{
	taskId: number,
	canWrite?: boolean
	initialComments: ITaskComment[]
}>(), {
	canWrite: true,
})

const copy = useCopyToClipboard()

const {t} = useI18n({useScope: 'global'})
const configStore = useConfigStore()
const authStore = useAuthStore()

const comments = ref<ITaskComment[]>([])

const showDeleteModal = ref(false)
const commentToDelete = reactive(new TaskCommentModel())

const isCommentEdit = ref(false)
const commentEdit = reactive(new TaskCommentModel())

const newCommentText = ref('')

const saved = ref<ITask['id'] | null>(null)
const saving = ref<ITask['id'] | null>(null)

const userAvatar = ref('')
const avatarCache = reactive(new Map<string, string>())

function avatarFor(u: IUser, size: number) {
	const key = `${u.id}-${size}`
	const cached = avatarCache.get(key)
	if (!cached) {
		fetchAvatarBlobUrl(u, size).then(url => avatarCache.set(key, url))
	}

	return avatarCache.get(key) || ''
}

watch(() => authStore.info, async (nu) => {
	if (!nu) {
		return
	}
	userAvatar.value = await fetchAvatarBlobUrl(nu, 48)
}, {immediate: true})

const currentUserId = computed(() => authStore.info.id)
const enabled = computed(() => configStore.taskCommentsEnabled)
const actions = computed(() => {
	if (!props.canWrite) {
		return {}
	}
	return Object.fromEntries(comments.value.map((comment) => ([
		comment.id,
		comment.author.id === currentUserId.value
			? [{
				action: () => toggleDelete(comment.id),
				title: t('misc.delete'),
			}]
			: [],
	])))
})

const frontendUrl = computed(() => configStore.frontendUrl)

const currentPage = ref(1)

const commentsRef = ref<HTMLElement | null>(null)

async function attachmentUpload(files: File[] | FileList): (Promise<string[]>) {

	const uploadPromises: Promise<string>[] = []

	files.forEach((file: File) => {
		const promise = new Promise<string>((resolve) => {
			uploadFile(props.taskId, file, (uploadedFileUrl: string) => resolve(uploadedFileUrl))
		})

		uploadPromises.push(promise)
	})

	return await Promise.all(uploadPromises)
}

const taskCommentService = shallowReactive(new TaskCommentService())

async function loadComments(taskId: ITask['id']) {
	if (!enabled.value) {
		return
	}

	commentEdit.taskId = taskId
	commentToDelete.taskId = taskId
	
	if(typeof props.initialComments !== 'undefined' && currentPage.value === 1) {
		comments.value = props.initialComments
		return
	}

	comments.value = await taskCommentService.getAll({taskId}, {}, currentPage.value)
}

async function changePage(page: number) {
	commentsRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start', inline: 'nearest' })
	currentPage.value = page
	await loadComments(props.taskId)
}

watch(
	() => props.taskId,
	() => {
		currentPage.value = 1 // Reset to first page when task changes
		loadComments(props.taskId)
	},
	{immediate: true},
)

const editorActive = ref(true)
const creating = ref(false)

async function addComment() {
	if (newCommentText.value === '') {
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
		const newComment = new TaskCommentModel()
		newComment.taskId = props.taskId
		newComment.comment = newCommentText.value
		const comment = await taskCommentService.create(newComment)
		comments.value.push(comment)
		newCommentText.value = ''
		success({message: t('task.comment.addedSuccess')})
	} finally {
		creating.value = false
	}
}

function toggleEdit(comment: ITaskComment) {
	isCommentEdit.value = !isCommentEdit.value
	Object.assign(commentEdit, comment)
}

function toggleDelete(commentId: ITaskComment['id']) {
	showDeleteModal.value = !showDeleteModal.value
	commentToDelete.id = commentId
}

const changeTimeout = ref<ReturnType<typeof setTimeout> | null>(null)

async function editCommentWithDelay() {
	if (changeTimeout.value !== null) {
		clearTimeout(changeTimeout.value)
	}

	changeTimeout.value = setTimeout(async () => {
		await editComment()
	}, 5000)
}

async function editComment() {
	if (commentEdit.comment === '') {
		return
	}

	if (changeTimeout.value !== null) {
		clearTimeout(changeTimeout.value)
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

async function deleteComment(commentToDelete: ITaskComment) {
	try {
		await taskCommentService.delete(commentToDelete)
		const index = comments.value.findIndex(({id}) => id === commentToDelete.id)
		comments.value.splice(index, 1)
		success({message: t('task.comment.deleteSuccess')})
	} finally {
		showDeleteModal.value = false
	}
}

function getCommentUrl(commentId: string) {
	const baseUrl = frontendUrl.value.endsWith('/') ? frontendUrl.value.slice(0, -1) : frontendUrl.value
	return `${baseUrl}${location.pathname}${location.search}#comment-${commentId}`
}
</script>

<style lang="scss" scoped>
.media-left {
	margin: 0 1rem !important;
}

.comment-info {
	display: flex;
	align-items: center;
	gap: .5rem;

	img {
		@media screen and (max-width: $tablet) {
			display: block;
			inline-size: 20px;
			block-size: 20px;
			padding-inline-end: 0;
			margin-inline-end: .5rem;
		}

		@media screen and (min-width: $tablet) {
			display: none;
		}
	}


	span,
	.comment-permalink {
		font-size: .75rem;
		line-height: 1;
	}

	.comment-permalink {
		font-size: 1rem;
		border: 1px solid transparent;
		padding: 0.25rem;
		border-radius: 1rem;
		color: var(--grey, hsl(0, 0%, 48%));
	}
	.comment-permalink:hover {
		color: var(--grey-dark, hsl(0, 0%, 29%));
		border-color: var(--grey-dark, hsl(0, 0%, 29%));
	}
}

.image.is-avatar {
	border-radius: 100%;
}

.media-content {
	inline-size: calc(100% - 48px - 2rem);
}

.comments-container {
	scroll-margin-block-start: 4rem;
}
</style>
