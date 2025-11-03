<script setup lang="ts">
import type {IReactionPerEntity, ReactionKind} from '@/modelTypes/IReaction'
import {VuemojiPicker} from 'vuemoji-picker'
import ReactionService from '@/services/reactions'
import ReactionModel from '@/models/reaction'
import BaseButton from '@/components/base/BaseButton.vue'
import type {IUser} from '@/modelTypes/IUser'
import {getDisplayName} from '@/models/user'
import {useI18n} from 'vue-i18n'
import {nextTick, onBeforeUnmount, onMounted, ref} from 'vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'
import {useAuthStore} from '@/stores/auth'
import {useColorScheme} from '@/composables/useColorScheme'

const props = withDefaults(defineProps<{
	entityKind: ReactionKind,
	entityId: number,
	disabled?: boolean,
}>(), {
	disabled: false,
})

const model = defineModel<IReactionPerEntity>()

const authStore = useAuthStore()
const {t} = useI18n()
const reactionService = new ReactionService()
const {isDark} = useColorScheme()

async function addReaction(value: string) {
	const reaction = new ReactionModel({
		id: props.entityId,
		kind: props.entityKind,
		value,
	})
	await reactionService.create(reaction)
	showEmojiPicker.value = false

	if (typeof model.value === 'undefined') {
		model.value = {}
	}

	if (typeof model.value[reaction.value] === 'undefined') {
		model.value[reaction.value] = [authStore.info]
	} else {
		model.value[reaction.value].push(authStore.info)
	}
}

async function removeReaction(value: string) {
	const reaction = new ReactionModel({
		id: props.entityId,
		kind: props.entityKind,
		value,
	})
	await reactionService.delete(reaction)
	showEmojiPicker.value = false
	
	const userIndex = model.value[reaction.value].findIndex(u => u.id === authStore.info?.id)
	if (userIndex !== -1) {
		model.value[reaction.value].splice(userIndex, 1)
	}
	if(model.value[reaction.value].length === 0) {
		delete model.value[reaction.value]
	}
}

function getReactionTooltip(users: IUser[], value: string) {
	const names = users.map(u => getDisplayName(u))

	if (names.length === 1) {
		return t('reaction.reactedWith', {user: names[0], value})
	}

	if (names.length > 1 && names.length < 10) {
		return t('reaction.reactedWithAnd', {
			users: names.slice(0, names.length - 1).join(', '),
			lastUser: names[names.length - 1],
			value,
		})
	}

	return t('reaction.reactedWithAndMany', {
		users: names.slice(0, 10).join(', '),
		num: names.length - 10,
		value,
	})
}

const showEmojiPicker = ref(false)
const emojiPickerRef = ref<HTMLElement | null>(null)

function hideEmojiPicker(e: MouseEvent) {
	if (showEmojiPicker.value) {
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
		const rect = emojiPickerButtonRef.value?.$el.getBoundingClientRect()
		const container = reactionContainerRef.value?.getBoundingClientRect()
		const left = rect.left - container.left + rect.width

		emojiPickerPosition.value = {
			'inset-inline-start': left === 0 ? undefined : left,
		}
	}

	nextTick(() => showEmojiPicker.value = !showEmojiPicker.value)
}

function hasCurrentUserReactedWithEmoji(value: string): boolean {
	const user = model.value[value].find(u => u.id === authStore.info.id)
	return typeof user !== 'undefined'
}

async function toggleReaction(value: string) {
	if (hasCurrentUserReactedWithEmoji(value)) {
		return removeReaction(value)
	}
	
	return addReaction(value)
}
</script>

<template>
	<div
		ref="reactionContainerRef"
		class="reactions"
	>
		<BaseButton
			v-for="(users, value) in (model as IReactionPerEntity)"
			:key="'button' + value"
			v-tooltip="getReactionTooltip(users, value)"
			class="reaction-button"
			:class="{'current-user-has-reacted': hasCurrentUserReactedWithEmoji(value)}"
			:disabled
			@click="toggleReaction(value)"
		>
			{{ value }} {{ users.length }}
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
