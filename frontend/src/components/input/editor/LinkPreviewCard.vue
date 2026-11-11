<template>
	<a
		v-if="preview && (preview.title || preview.description || image)"
		class="link-preview"
		:href="href"
		target="_blank"
		rel="noopener noreferrer nofollow"
	>
		<img
			v-if="image && failedImage !== image"
			class="link-preview__image"
			:src="image"
			alt=""
			loading="lazy"
			referrerpolicy="no-referrer"
			@error="failedImage = image"
		>
		<span class="link-preview__body">
			<span class="link-preview__site">
				<img
					v-if="favicon && failedFavicon !== favicon"
					:src="favicon"
					alt=""
					loading="lazy"
					referrerpolicy="no-referrer"
					@error="failedFavicon = favicon"
				>
				{{ preview.site_name || hostname }}
			</span>
			<strong class="link-preview__title">{{ preview.title || href }}</strong>
			<span
				v-if="preview.description"
				class="link-preview__description"
			>{{ preview.description }}</span>
		</span>
	</a>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {linkPreviewQuery} from '@/client/queries/linkPreviews'
import {safePreviewUrl} from '@/helpers/standaloneLinkUrls'

const props = defineProps<{href: string}>()
const {data: preview} = useQuery(computed(() => linkPreviewQuery(props.href)))
const hostname = computed(() => safePreviewUrl(props.href)?.hostname ?? '')
const image = computed(() => safePreviewUrl(preview.value?.image ?? '')?.href ?? '')
const favicon = computed(() => safePreviewUrl(preview.value?.favicon ?? '')?.href ?? '')
const failedImage = ref('')
const failedFavicon = ref('')
</script>

<style scoped lang="scss">
.link-preview {
	display: flex;
	overflow: hidden;
	max-inline-size: 36rem;
	margin-block: .75rem;
	border: 1px solid var(--grey-200);
	border-inline-start: 3px solid var(--primary);
	border-radius: $radius;
	background: var(--white);
	color: var(--text);
	text-decoration: none;

	&:hover {
		border-color: var(--primary);
		text-decoration: none;
	}

	&:focus-visible {
		outline: 2px solid var(--primary);
		outline-offset: 2px;
	}

	&__image {
		flex: 0 0 6rem;
		inline-size: 6rem;
		max-block-size: 9rem;
		object-fit: cover;
	}

	&__body {
		display: flex;
		flex-direction: column;
		gap: .25rem;
		min-inline-size: 0;
		padding: .75rem;
		overflow-wrap: anywhere;
	}

	&__site {
		display: flex;
		align-items: center;
		gap: .375rem;
		color: var(--grey-600);
		font-size: .875rem;

		img {
			inline-size: 1rem;
			block-size: 1rem;
			object-fit: contain;
		}
	}

	&__title {
		color: var(--link);
	}

	&__description {
		display: -webkit-box;
		overflow: hidden;
		font-size: .875rem;
		-webkit-box-orient: vertical;
		-webkit-line-clamp: 3;
	}
}
</style>
