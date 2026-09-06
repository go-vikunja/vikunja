<template>
	<img
		v-if="src"
		:src="src"
		:alt="alt"
		:width="size"
		:height="size"
	>
	<span
		v-else
		:style="{'--user-avatar-size': `${size}px`}"
		class="user-avatar-placeholder"
		aria-hidden="true"
	/>
</template>

<script lang="ts" setup>
import {ref, watch} from 'vue'

import {avatarCacheVersions, fetchAvatarBlobUrl} from '@/models/user'
import type {IUser} from '@/modelTypes/IUser'

const props = withDefaults(defineProps<{
	user?: Pick<IUser, 'username'> | null,
	size?: number,
	// Empty by default: most avatars sit next to a visible name and are decorative.
	alt?: string,
}>(), {
	user: undefined,
	size: 50,
	alt: '',
})

const src = ref<string>()

// Guards against a slow fetch for a previous user overwriting a newer one.
let fetchToken = 0

watch(
	[() => props.user?.username, () => props.size, () => avatarCacheVersions.get(props.user?.username ?? '')],
	async () => {
		const token = ++fetchToken
		src.value = undefined

		if (!props.user?.username) {
			return
		}

		try {
			const url = await fetchAvatarBlobUrl(props.user, props.size)
			if (token === fetchToken) {
				src.value = url
			}
		} catch {
			// A missing avatar isn't worth a user-visible error; used to end up in Sentry unhandled.
		}
	},
	{immediate: true},
)
</script>

<style lang="scss">
// Unscoped so call site styles win over the fallback size without !important or :deep().
.user-avatar-placeholder {
	display: inline-block;
	inline-size: var(--user-avatar-size);
	block-size: var(--user-avatar-size);
}
</style>
