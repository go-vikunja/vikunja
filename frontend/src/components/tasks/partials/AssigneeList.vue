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
				:key="'user' + user.id"
				:avatar-size="avatarSize"
				:show-username="true"
				:user="user"
				:class="{'m-2': canRemove && !disabled}"
			/>
			<BaseButton
				v-if="canRemove && !disabled"
				:key="'delete' + user.id"
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
	align-items: center;

	&:not(.is-inline) {
		flex-wrap: wrap;
		gap: 0.5rem;
		
		.assignee {
			position: relative;
			margin-left: 0;
			
			:deep(.user img) {
				border: 2px solid var(--white);
				margin-right: 0.25rem;
			}
			
			:deep(.user .username) {
				font-size: 0.85rem;
				display: inline-block;
				max-width: 100px;
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;
				vertical-align: middle;
			}
		}
	}

	&.is-inline {
		flex-wrap: nowrap;
		overflow: hidden;
		
		:deep(.user) {
			display: inline-flex;
			align-items: center;
		}

		.assignee {
			position: relative;
			display: inline-flex;
			align-items: center;
			max-width: fit-content;
			margin-right: 0.25rem;
			
			:deep(.user img) {
				border: 2px solid var(--white);
				margin-right: 0.25rem;
				z-index: 1;
			}
			
			:deep(.user .username) {
				font-size: 0.75rem;
				display: inline-block;
				max-width: 50px;
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;
				vertical-align: middle;
				opacity: 0.85;
			}
		}
	}
}

.assignee {
	position: relative;
	margin-right: 0.25rem;

	:deep(.user img) {
		border: 2px solid var(--white);
		margin-right: 0.25rem;
	}
}

.remove-assignee {
	position: absolute;
	top: 4px;
	left: 2px;
	color: var(--danger);
	background: var(--white);
	padding: 0 4px;
	display: block;
	border-radius: 100%;
	font-size: .75rem;
	width: 18px;
	height: 18px;
	z-index: 100;
}
</style>