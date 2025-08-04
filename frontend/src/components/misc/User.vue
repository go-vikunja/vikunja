<template>
	<div
		class="user"
		:class="{'is-inline': isInline}"
	>
		<img
			v-tooltip="displayName"
			:height="avatarSize"
			:src="avatarSrc"
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
import {computed, ref, watch} from 'vue'

import {fetchAvatarBlobUrl, getDisplayName} from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

const props = withDefaults(defineProps<{
	user: IUser,
	showUsername?: boolean,
	avatarSize?: number,
	isInline?: boolean,
}>(), {
	showUsername: true,
	avatarSize: 50,
	isInline: false,
})

const displayName = computed(() => getDisplayName(props.user))
const avatarSrc = ref('')

async function loadAvatar() {
	avatarSrc.value = await fetchAvatarBlobUrl(props.user, props.avatarSize)
}

watch(() => [props.user, props.avatarSize], loadAvatar, { immediate: true })
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
	margin-inline-end: .5rem;
}
</style>
