<template>
	<NodeViewWrapper
		as="blockquote"
		class="comment-quote"
		:class="{'comment-quote--has-parent': hasParent}"
		:data-comment-id="commentId === null ? null : String(commentId)"
	>
		<div
			v-if="commentId !== null && ctx"
			contenteditable="false"
			class="comment-quote__header"
		>
			<template v-if="parent">
				<img
					v-if="avatarUrl"
					:src="avatarUrl"
					alt=""
					class="comment-quote__avatar"
					width="20"
					height="20"
				>
				<span class="comment-quote__author">{{ authorName }}</span>
				<BaseButton
					v-tooltip="t('task.comment.jumpToOriginal')"
					class="comment-quote__jump"
					:aria-label="t('task.comment.jumpToOriginal')"
					@click="onJump"
				>
					<Icon icon="angle-right" />
				</BaseButton>
			</template>
			<span
				v-else
				class="comment-quote__author comment-quote__author--missing"
			>
				{{ t('task.comment.deletedComment') }}
			</span>
		</div>
		<NodeViewContent class="comment-quote__body" />
	</NodeViewWrapper>
</template>

<script lang="ts" setup>
import {computed, inject, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {nodeViewProps, NodeViewWrapper, NodeViewContent} from '@tiptap/vue-3'

import BaseButton from '@/components/base/BaseButton.vue'
import {fetchAvatarBlobUrl, getDisplayName} from '@/models/user'
import {commentReplyContextKey} from '@/components/tasks/partials/commentReplyContext'

const props = defineProps(nodeViewProps)

const {t} = useI18n({useScope: 'global'})

const ctx = inject(commentReplyContextKey, null)

const commentId = computed<number | null>(() => {
	const raw = props.node.attrs.commentId
	if (raw === null || raw === undefined) {
		return null
	}
	const id = Number(raw)
	return Number.isInteger(id) && id > 0 ? id : null
})

const parent = computed(() => {
	if (commentId.value === null || !ctx) {
		return undefined
	}
	return ctx.findComment(commentId.value)
})

const hasParent = computed(() => parent.value !== undefined)

const authorName = computed(() => {
	const p = parent.value
	return p ? getDisplayName(p.author) : ''
})

const avatarUrl = ref('')

// Bumped on every parent change so stale avatar fetches (older parent)
// don't overwrite a newer one if the user navigates between comments
// while fetches are still in flight.
let avatarFetchToken = 0

watch(parent, (p) => {
	avatarUrl.value = ''
	const token = ++avatarFetchToken
	if (!p?.author) {
		return
	}
	fetchAvatarBlobUrl(p.author, 20)
		.then((url) => {
			if (token === avatarFetchToken) {
				avatarUrl.value = (url as string) ?? ''
			}
		})
		.catch(() => {
			// Swallow — a missing avatar isn't worth a user-visible error;
			// the header still renders with the author name.
		})
}, {immediate: true})

function onJump() {
	if (commentId.value !== null && ctx) {
		ctx.scrollToComment(commentId.value)
	}
}
</script>

<style lang="scss">
.tiptap blockquote.comment-quote {
	margin-block: .5rem;

	.comment-quote__header {
		display: flex;
		align-items: center;
		gap: .5rem;
		padding-block-end: .25rem;
		font-size: .85rem;
		color: var(--grey-600);
		user-select: none;
	}

	.comment-quote__avatar {
		border-radius: 50%;
		flex: 0 0 auto;
	}

	.comment-quote__author {
		font-weight: 600;
		color: var(--grey-700);

		&--missing {
			font-style: italic;
			color: var(--grey-500);
		}
	}

	.comment-quote__jump {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		color: var(--grey-500);
		padding: .15rem .25rem;
		border-radius: 9999px;
		transition: background-color $transition, color $transition;

		&:hover {
			color: var(--grey-800);
			background: var(--grey-200);
		}
	}

	.comment-quote__body > :first-child {
		margin-block-start: 0;
	}
}
</style>
