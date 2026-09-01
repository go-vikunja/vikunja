<template>
	<div class="unsplash-thumbnail">
		<BaseButton
			class="unsplash-thumbnail__button"
			:aria-label="author?.author_name ? $t('project.background.setBackgroundBy', {author: author.author_name}) : $t('project.background.setBackground')"
			:style="blurHashUrl ? {'background-image': `url(${blurHashUrl})`} : undefined"
			@click="emit('select')"
		>
			<CustomTransition name="fade">
				<img
					v-if="thumbUrl"
					class="unsplash-thumbnail__image"
					:src="thumbUrl"
					alt=""
				>
			</CustomTransition>
		</BaseButton>

		<BaseButton
			v-if="author"
			:href="`https://unsplash.com/@${encodeURIComponent(author.author)}?utm_source=vikunja&utm_medium=referral`"
			class="unsplash-thumbnail__info"
		>
			{{ author.author_name }}
		</BaseButton>
	</div>
</template>

<script setup lang="ts">
import {computed, watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useObjectUrl} from '@vueuse/core'

import type {Image} from '@/client/generated'
import {
	unsplashAuthor,
	unsplashBackgroundThumbnailQuery,
} from '@/client/queries/projectBackgrounds'
import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import {useBlurHashUrl} from '@/composables/useBlurHashUrl'

const props = defineProps<{
	image: Image & {id: string},
}>()

const emit = defineEmits<{
	select: [],
}>()

const author = computed(() => unsplashAuthor(props.image.info))

const thumb = useQuery(computed(() => unsplashBackgroundThumbnailQuery(props.image.id)))
const thumbUrl = useObjectUrl(thumb.data)

watch(thumb.error, error => {
	if (error !== null) {
		console.error('Failed to load Unsplash thumbnail', props.image.id, error instanceof Error ? error.message : error)
	}
})

const blurHashUrl = useBlurHashUrl(() => props.image.blur_hash ?? '')
</script>

<style lang="scss" scoped>
.unsplash-thumbnail {
	position: relative;
	display: flex;
	inline-size: 100%;
}

.unsplash-thumbnail__button {
	inline-size: 100%;
	background-size: cover;
	background-position: center;
}

.unsplash-thumbnail__image {
	inline-size: 100%;
	block-size: 100%;
	object-fit: cover;
}

.unsplash-thumbnail__info {
	position: absolute;
	inset-block-end: 0;
	inline-size: 100%;
	padding: .25rem 0;
	opacity: 0;
	text-align: center;
	background: rgba(0, 0, 0, 0.5);
	font-size: .75rem;
	font-weight: bold;
	color: $white;
	transition: opacity $transition;

	&:focus-visible {
		opacity: 1;
	}
}

.unsplash-thumbnail:hover .unsplash-thumbnail__info,
.unsplash-thumbnail:focus-within .unsplash-thumbnail__info {
	opacity: 1;
}
</style>
