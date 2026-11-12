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
})))
// An svg avatar arrives as an inert data: url, everything else as bytes this component owns a url for.
const bytes = computed(() => avatar.data.value instanceof Blob ? avatar.data.value : undefined)
const objectUrl = useObjectUrl(bytes)
const src = computed(() => typeof avatar.data.value === 'string' ? avatar.data.value : objectUrl.value)
</script>

<style lang="scss">
// Unscoped so call site styles win over the fallback size without !important or :deep().
.user-avatar-placeholder {
	display: inline-block;
	inline-size: var(--user-avatar-size);
	block-size: var(--user-avatar-size);
}
</style>
