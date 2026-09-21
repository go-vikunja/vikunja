<script setup lang="ts">
import type {ReactionInput, ReactionUsers} from '@/client/queries/reactions'
import {VuemojiPicker} from 'vuemoji-picker'
import {useSetReactionMutation} from '@/client/queries/reactions'
import BaseButton from '@/components/base/BaseButton.vue'
import {getDisplayName} from '@/helpers/user'
import {useI18n} from 'vue-i18n'
import {nextTick, onBeforeUnmount, onMounted, ref} from 'vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {useAuthStore} from '@/stores/auth'
import {useColorScheme} from '@/composables/useColorScheme'

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
const {isDark} = useColorScheme()

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
	showEmojiPicker.value = false
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
const emojiPickerRef = ref<InstanceType<typeof VuemojiPicker> | null>(null)

function hideEmojiPicker(e: MouseEvent) {
	if (showEmojiPicker.value && emojiPickerRef.value?.$el) {
		closeWhenClickedOutside(e, emojiPickerRef.value.$el, () => showEmojiPicker.value = false)
	}
}

onMounted(() => document.addEventListener('click', hideEmojiPicker))
onBeforeUnmount(() => document.removeEventListener('click', hideEmojiPicker))

const emojiPickerButtonRef = ref<HTMLElement | null>(null)
const reactionContainerRef = ref<HTMLElement | null>(null)
const emojiPickerPosition = ref()

function toggleEmojiPicker() {
	if (!showEmojiPicker.value) {
		const rect = emojiPickerButtonRef.value?.$el?.getBoundingClientRect()
		const container = reactionContainerRef.value?.getBoundingClientRect()
		if (rect && container) {
			const left = rect.left - container.left + rect.width

			emojiPickerPosition.value = {
				'inset-inline-start': left === 0 ? undefined : left,
			}
		}
	}

	nextTick(() => showEmojiPicker.value = !showEmojiPicker.value)
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
	<div
		ref="reactionContainerRef"
		class="reactions"
	>
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
		<BaseButton
			v-if="!disabled"
			ref="emojiPickerButtonRef"
			v-tooltip="$t('reaction.add')"
			class="reaction-button"
			@click.stop="toggleEmojiPicker"
		>
			<span class="is-sr-only">{{ $t('reaction.add') }}</span>
			<Icon :icon="['far', 'face-laugh']" />
		</BaseButton>
		<CustomTransition name="fade">
			<VuemojiPicker
				v-if="showEmojiPicker"
				ref="emojiPickerRef"
				class="emoji-picker"
				:style="{'inset-inline-start': emojiPickerPosition?.left + 'px'}"
				data-source="/emojis.json"
				:is-dark="isDark"
				@emojiClick="detail => addReaction(detail.unicode)"
			/>
		</CustomTransition>
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

.emoji-picker {
	position: absolute;
	z-index: 99;
	margin-block-start: .5rem;
}
</style>
