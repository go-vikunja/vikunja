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
			v-if="showUsername"
			class="username"
		>{{ displayName }}</span>
	</div>
</template>

<script lang="ts" setup>
import {computed, type PropType} from 'vue'

import {getAvatarUrl, getDisplayName} from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

const props = defineProps({
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
