<template>
	<BaseButton
		class="image-search__image-button"
		:style="blurHashUrl ? {'background-image': `url(${blurHashUrl})`} : undefined"
		@click="emit('select')"
	>
		<CustomTransition name="fade">
			<img
				v-if="thumbUrl"
				class="image-search__image"
				:src="thumbUrl"
				alt=""
			>
		</CustomTransition>
	</BaseButton>
</template>

<script setup lang="ts">
import {computed, shallowRef, watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useObjectUrl} from '@vueuse/core'

import {
	unsplashBackgroundThumbnailQuery,
	type UnsplashSearchImage,
} from '@/client/queries/projectBackgrounds'
import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

const props = defineProps<{
	image: UnsplashSearchImage,
}>()

const emit = defineEmits<{
	select: [],
}>()

const thumb = useQuery(computed(() => unsplashBackgroundThumbnailQuery(props.image.id)))
const thumbUrl = useObjectUrl(thumb.data)

watch(thumb.error, error => {
	if (error !== null) {
		console.error(error)
	}
})

const blurBlob = shallowRef<Blob | null>(null)
watch(() => props.image.blur_hash, (blurHash, _previous, onCleanup) => {
	let current = true
	onCleanup(() => {
		current = false
	})

	blurBlob.value = null
	getBlobFromBlurHash(blurHash)
		.then(blob => {
			if (current) {
				blurBlob.value = blob
			}
		})
		.catch(error => console.error(error))
}, {immediate: true})
const blurHashUrl = useObjectUrl(blurBlob)
</script>

<style lang="scss" scoped>
.image-search__image-button {
	inline-size: 100%;
	background-size: cover;
	background-position: center;
}

.image-search__image {
	inline-size: 100%;
	block-size: 100%;
	object-fit: cover;
}
</style>
