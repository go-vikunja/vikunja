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
import {computed} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useObjectUrl} from '@vueuse/core'
import {avatarQuery} from '@/client/queries/avatars'
import {queryClient} from '@/client/queryClient'
import type {User as IUser} from '@/client/generated'

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

const avatar = useQuery(computed(() => ({
	...avatarQuery(props.user?.username ?? '', props.size),
	enabled: Boolean(props.user?.username),
})), queryClient)
const src = useObjectUrl(computed(() => props.user?.username ? avatar.data.value : undefined))
</script>

<style lang="scss">
// Unscoped so call site styles win over the fallback size without !important or :deep().
.user-avatar-placeholder {
	display: inline-block;
	inline-size: var(--user-avatar-size);
	block-size: var(--user-avatar-size);
}
</style>
