<script setup lang="ts">
import type {IUser} from '@/modelTypes/IUser'
import BaseButton from '@/components/base/BaseButton.vue'
import User from '@/components/misc/User.vue'

withDefaults(defineProps<{
	assignees: IUser[],
	disabled?: boolean,
	avatarSize?: number,
	inline?: boolean,
	/** add this boolean prop to enable removal of assignees */
	canRemove?: boolean,
}>(), {
	avatarSize: 30,
	inline: false,
	canRemove: false,
})

defineEmits<{
	remove: [user: IUser],
}>()
</script>

<template>
	<div
		class="assignees-list"
		:class="{'is-inline': inline}"
	>
		<span
			v-for="user in assignees"
			:key="user.id"
			class="assignee"
		>
			<User
				:key="'user'+user.id"
				:avatar-size="avatarSize"
				:show-username="false"
				:user="user"
				:class="{'m-2': canRemove && !disabled}"
			/>
			<BaseButton
				v-if="canRemove && !disabled"
				:key="'delete'+user.id"
				class="remove-assignee"
				@click="$emit('remove', user)"
			>
				<Icon icon="times" />
			</BaseButton>
		</span>
	</div>
</template>

<style scoped lang="scss">
.assignees-list {
	display: flex;

	&.is-inline :deep(.user) {
		display: inline;
	}

	&:hover .assignee:not(:first-child) {
		margin-inline-start: -0.5rem;
	}
}

.assignee {
	position: relative;
	transition: all $transition;

	&:not(:first-child) {
		margin-inline-start: -1rem;
	}

	:deep(.user img) {
		border: 2px solid var(--white);
		margin-inline-end: 0;
	}
}

.remove-assignee {
	position: absolute;
	inset-block-start: 4px;
	inset-inline-start: 2px;
	color: var(--danger);
	background: var(--white);
	padding: 0 4px;
	display: block;
	border-radius: 100%;
	font-size: .75rem;
	inline-size: 18px;
	block-size: 18px;
	z-index: 100;
}
</style>
