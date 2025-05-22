<template>
	<div
		class="user"
		:class="{'is-inline': isInline}"
	>
		<img
			v-tooltip="displayName"
			:height="avatarSize"
			:src="getAvatarUrl(user, avatarSize)"
			:width="avatarSize"
			:alt="'Avatar of ' + displayName"
			class="avatar"
		>
		<span
			v-if="!hideUsername"
			class="username"
		>{{ displayName }}</span>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'

import {getAvatarUrl, getDisplayName} from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

const props = withDefaults(defineProps<{
	user: IUser,
	hideUsername?: boolean,
	avatarSize?: number,
	isInline?: boolean,
}>(), {
	hideUsername: false,
	avatarSize: 50,
	isInline: false,
})

const displayName = computed(() => getDisplayName(props.user))
</script>

<style lang="scss" scoped>
.user {
	display: flex;
	justify-items: center;

	&.is-inline {
		display: inline-flex;
	}
}

.avatar {
	border-radius: 100%;
	vertical-align: middle;
	margin-right: .5rem;
}
</style>
