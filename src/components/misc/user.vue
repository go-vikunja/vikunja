<template>
	<div :class="{'is-inline': isInline}" class="user">
		<img
			:height="avatarSize"
			:src="getAvatarUrl(user, avatarSize)"
			:width="avatarSize"
			alt=""
			class="avatar"
			v-tooltip="getDisplayName(user)"/>
		<span class="username" v-if="showUsername">{{ getDisplayName(user) }}</span>
	</div>
</template>

<script lang="ts" setup>
import type {PropType} from 'vue'

import {getAvatarUrl, getDisplayName} from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

defineProps({
	user: {
		type: Object as PropType<IUser>,
		required: true,
	},
	showUsername: {
		type: Boolean,
		required: false,
		default: true,
	},
	avatarSize: {
		type: Number,
		required: false,
		default: 50,
	},
	isInline: {
		type: Boolean,
		required: false,
		default: false,
	},
})
</script>

<style lang="scss" scoped>
.user {
	margin: .5rem;

	&.is-inline {
		display: inline;
	}

	img {
		border-radius: 100%;

		vertical-align: middle;
		margin-right: .5rem;
	}
}
</style>
