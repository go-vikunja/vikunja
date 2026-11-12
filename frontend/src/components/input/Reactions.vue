<script setup lang="ts">
import type {ReactionInput, ReactionUsers} from '@/client/queries/reactions'
import EmojiPicker from '@/components/input/EmojiPicker.vue'
import {useSetReactionMutation} from '@/client/queries/reactions'
import BaseButton from '@/components/base/BaseButton.vue'
import {getDisplayName} from '@/helpers/user'
import {useI18n} from 'vue-i18n'
import {nextTick, ref, useId} from 'vue'
import Popup from '@/components/misc/Popup.vue'
import {useAuthStore} from '@/stores/auth'

type ReactionSubject = {
	entityKind: 'tasks',
	taskId?: never,
} | {
	entityKind: 'comments',
	taskId: number,
}

const props = withDefaults(defineProps<ReactionSubject & {
	entityId: number,
	modelValue?: ReactionUsers,
	disabled?: boolean,
}>(), {
	modelValue: undefined,
	disabled: false,
})

const authStore = useAuthStore()
const {t} = useI18n()
const reactionMutation = useSetReactionMutation()

async function setReaction(value: string, remove: boolean) {
	if (props.disabled || reactionMutation.isPending.value || !authStore.info) return
	const user = {
		id: authStore.info.id,
		name: authStore.info.name,
		username: authStore.info.username,
	}
	const input: ReactionInput = props.entityKind === 'comments'
		? {
			kind: 'comments',
			taskId: props.taskId,
			id: props.entityId,
			value,
			remove,
			user,
		}
		: {
			kind: 'tasks',
			id: props.entityId,
			value,
			remove,
			user,
		}
	try {
		await reactionMutation.mutateAsync(input)
	} catch {
		return
	}
	if (props.entityId !== input.id || props.entityKind !== input.kind) return
	if (showEmojiPicker.value) closeEmojiPicker()
}

function addReaction(value: string) {
	return setReaction(value, false)
}

function removeReaction(value: string) {
	return setReaction(value, true)
}

function getReactionTooltip(users: ReactionUsers[string], value: string | number) {
	const names = (users ?? []).map(u => getDisplayName(u))
	const valueStr = String(value)

	if (names.length === 1) {
		return t('reaction.reactedWith', {user: names[0], value: valueStr})
	}

	if (names.length > 1 && names.length < 10) {
		return t('reaction.reactedWithAnd', {
			users: names.slice(0, names.length - 1).join(', '),
			lastUser: names[names.length - 1],
			value: valueStr,
		})
	}

	return t('reaction.reactedWithAndMany', {
		users: names.slice(0, 10).join(', '),
		num: names.length - 10,
		value: valueStr,
	})
}

const showEmojiPicker = ref(false)
const emojiPickerButtonRef = ref<InstanceType<typeof BaseButton> | null>(null)
const pickerId = useId()

async function closeEmojiPicker() {
	showEmojiPicker.value = false
	await nextTick()
	emojiPickerButtonRef.value?.$el?.focus()
}

function hasCurrentUserReactedWithEmoji(value: string | number): boolean {
	if (!props.modelValue || !authStore.info) return false
	const user = props.modelValue[String(value)]?.find(u => u.id === authStore.info!.id)
	return typeof user !== 'undefined'
}

async function toggleReaction(value: string | number) {
	const valueStr = String(value)
	if (hasCurrentUserReactedWithEmoji(valueStr)) {
		return removeReaction(valueStr)
	}

	return addReaction(valueStr)
}
</script>

<template>
	<div class="reactions">
		<BaseButton
			v-for="(users, value) in modelValue"
			:key="'button' + value"
			v-tooltip="getReactionTooltip(users, value)"
			class="reaction-button"
			:class="{'current-user-has-reacted': hasCurrentUserReactedWithEmoji(value)}"
			:disabled="disabled"
			:aria-disabled="reactionMutation.isPending.value || undefined"
			:aria-pressed="hasCurrentUserReactedWithEmoji(value)"
			@click="toggleReaction(value)"
		>
			{{ value }} {{ users?.length }}
		</BaseButton>
		<Popup
			v-if="!disabled"
			v-model:open="showEmojiPicker"
			:anchor="emojiPickerButtonRef?.$el"
			placement="bottom-start"
		>
			<template #trigger="{toggle}">
				<BaseButton
					ref="emojiPickerButtonRef"
					v-tooltip="$t('reaction.add')"
					class="reaction-button"
					:aria-expanded="showEmojiPicker"
					:aria-controls="showEmojiPicker ? pickerId : undefined"
					@click.stop="toggle"
				>
					<span class="is-sr-only">{{ $t('reaction.add') }}</span>
					<Icon :icon="['far', 'face-laugh']" />
				</BaseButton>
			</template>
			<template #content>
				<EmojiPicker
					v-if="showEmojiPicker"
					:id="pickerId"
					@select="addReaction"
					@keydown.esc.stop.prevent="closeEmojiPicker"
				/>
			</template>
		</Popup>
	</div>
</template>

<style scoped lang="scss">
.reaction-button {
	margin-inline-end: .25rem;
	margin-block-end: .25rem;
	padding: .175rem .5rem .15rem;
	border: 1px solid var(--grey-400);
	background: var(--grey-100);
	border-radius: 100px;
	font-size: .75rem;

	&.current-user-has-reacted {
		border-color: var(--primary);
		background-color: hsla(var(--primary-h), var(--primary-s), var(--primary-light-l), 0.5);
	}
}
</style>
