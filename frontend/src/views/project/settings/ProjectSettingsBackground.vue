<template>
	<CreateEdit
		v-if="uploadBackgroundEnabled || unsplashBackgroundEnabled"
		:title="$t('project.background.title')"
		:loading="backgroundMutationLoading"
		class="project-background-setting"
		:wide="true"
	>
		<div
			v-if="uploadBackgroundEnabled"
			class="mbe-4"
		>
			<input
				ref="backgroundUploadInput"
				accept="image/*"
				class="is-hidden"
				type="file"
				@change="uploadBackground"
			>
			<XButton
				:loading="uploadBackgroundMutation.isPending.value"
				variant="primary"
				@click="backgroundUploadInput?.click()"
			>
				{{ $t('project.background.upload') }}
			</XButton>
		</div>
		<template v-if="unsplashBackgroundEnabled">
			<input
				v-model="backgroundSearchInput"
				:class="{'is-loading': backgroundSearchQuery.isFetching.value}"
				class="input is-expanded"
				:placeholder="$t('project.background.searchPlaceholder')"
				type="text"
				@keyup="debounceNewBackgroundSearch()"
			>

			<p class="unsplash-credit">
				<BaseButton
					class="unsplash-credit__link"
					href="https://unsplash.com"
				>
					{{ $t('project.background.poweredByUnsplash') }}
				</BaseButton>
			</p>

			<ul class="image-search__result-list">
				<li
					v-for="im in backgroundSearchResult"
					:key="im.id"
					class="image-search__result-item"
					:style="{'background-image': `url(${backgroundBlurHashes[im.id]})`}"
				>
					<CustomTransition name="fade">
						<BaseButton
							v-if="backgroundThumbs[im.id]"
							class="image-search__image-button"
							@click="setBackground(im.id)"
						>
							<img
								class="image-search__image"
								:src="backgroundThumbs[im.id]"
								alt=""
							>
						</BaseButton>
					</CustomTransition>

					<BaseButton
						:href="`https://unsplash.com/@${backgroundAuthors[im.id]?.author ?? ''}?utm_source=vikunja&utm_medium=referral`"
						class="image-search__info"
					>
						{{ backgroundAuthors[im.id]?.author_name ?? '' }}
					</BaseButton>
				</li>
			</ul>
			<XButton
				v-if="backgroundSearchResult.length > 0 && backgroundSearchQuery.hasNextPage.value"
				:disabled="backgroundSearchQuery.isFetchingNextPage.value"
				class="is-load-more-button mbs-4"
				:shadow="false"
				variant="secondary"
				@click="backgroundSearchQuery.fetchNextPage()"
			>
				{{ backgroundSearchQuery.isFetchingNextPage.value ? $t('misc.loading') : $t('project.background.loadMore') }}
			</XButton>
		</template>

		<template #footer>
			<XButton
				v-if="hasBackground"
				:shadow="false"
				variant="tertiary"
				danger
				@click.prevent.stop="removeBackground"
			>
				{{ $t('project.background.remove') }}
			</XButton>
			<XButton
				variant="secondary"
				@click.prevent.stop="$router.back()"
			>
				{{ $t('misc.close') }}
			</XButton>
		</template>
	</CreateEdit>
</template>


<script setup lang="ts">
import {computed, onBeforeUnmount, ref, watch, type Ref} from 'vue'
import {useInfiniteQuery, useQuery} from '@tanstack/vue-query'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'
import {useDebounceFn} from '@vueuse/core'

import type {Image} from '@/client/generated'
import {
	getUnsplashAuthorInfo,
	unsplashBackgroundSearchQuery,
	unsplashBackgroundThumbnailQuery,
	useDeleteProjectBackgroundMutation,
	useSetUnsplashProjectBackgroundMutation,
	useUploadProjectBackgroundMutation,
} from '@/client/queries/projectBackgrounds'
import {normalizeProject, projectQuery} from '@/client/queries/projects'
import {queryClient} from '@/client/queryClient'
import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import {useConfigStore} from '@/stores/config'
import {useBaseStore} from '@/stores/base'

import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'
import {useTitle} from '@/composables/useTitle'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import {success} from '@/message'

defineOptions({name: 'ProjectSettingBackground'})

const SEARCH_DEBOUNCE = 300

const {t} = useI18n({useScope: 'global'})
const route = useRoute()
const router = useRouter()
const configStore = useConfigStore()
const baseStore = useBaseStore()

useTitle(() => t('project.background.title'))

const unsplashBackgroundEnabled = computed(() => configStore.enabledBackgroundProviders.includes('unsplash'))
const uploadBackgroundEnabled = computed(() => configStore.enabledBackgroundProviders.includes('upload'))
const projectId = computed(() => Number(route.params.projectId))
const project = useQuery(computed(() => projectQuery(projectId.value)))
const backgroundSearchTerm = ref('')
const backgroundSearchInput = ref('')
const backgroundSearchQuery = useInfiniteQuery(computed(() => ({
	...unsplashBackgroundSearchQuery(backgroundSearchTerm.value),
	enabled: unsplashBackgroundEnabled.value,
})))
type SearchImage = Image & {id: string}
const backgroundSearchResult = computed(() =>
	(backgroundSearchQuery.data.value?.pages.flat() ?? [])
		.filter((image): image is SearchImage => typeof image.id === 'string'),
)
const backgroundThumbs = ref<Record<string, string>>({})
const backgroundBlurHashes = ref<Record<string, string>>({})

const hasBackground = computed(() => Boolean(project.data.value?.background_information))
const backgroundAuthors = computed(() => Object.fromEntries(
	backgroundSearchResult.value.map(image => [image.id, getUnsplashAuthorInfo(image)]),
))

const debounceNewBackgroundSearch = useDebounceFn(() => {
	backgroundSearchTerm.value = backgroundSearchInput.value
}, SEARCH_DEBOUNCE)

const setBackgroundMutation = useSetUnsplashProjectBackgroundMutation()
const uploadBackgroundMutation = useUploadProjectBackgroundMutation()
const deleteBackgroundMutation = useDeleteProjectBackgroundMutation()
const backgroundMutationLoading = computed(() =>
	setBackgroundMutation.isPending.value ||
	uploadBackgroundMutation.isPending.value ||
	deleteBackgroundMutation.isPending.value,
)

let mounted = true
let visibleImageIds = new Set<string>()
const pendingBlurHashes = new Set<string>()
const pendingThumbnails = new Set<string>()

function replaceObjectUrl(target: Ref<Record<string, string>>, imageId: string, blob: Blob) {
	const previous = target.value[imageId]
	if (previous) {
		window.URL.revokeObjectURL(previous)
	}
	target.value[imageId] = window.URL.createObjectURL(blob)
}

function removeObjectUrl(target: Ref<Record<string, string>>, imageId: string) {
	const url = target.value[imageId]
	if (!url) {
		return
	}
	window.URL.revokeObjectURL(url)
	delete target.value[imageId]
}

function removeInvisibleObjectUrls(target: Ref<Record<string, string>>) {
	Object.keys(target.value).forEach(imageId => {
		if (!visibleImageIds.has(imageId)) {
			removeObjectUrl(target, imageId)
		}
	})
}

watch(backgroundSearchResult, images => {
	visibleImageIds = new Set(images.map(image => image.id))
	removeInvisibleObjectUrls(backgroundThumbs)
	removeInvisibleObjectUrls(backgroundBlurHashes)

	images.forEach(image => {
		if (image.blur_hash && !backgroundBlurHashes.value[image.id] && !pendingBlurHashes.has(image.id)) {
			pendingBlurHashes.add(image.id)
			getBlobFromBlurHash(image.blur_hash)
				.then(blob => {
					if (mounted && visibleImageIds.has(image.id) && blob) {
						replaceObjectUrl(backgroundBlurHashes, image.id, blob)
					}
				})
				.catch(() => {})
				.finally(() => pendingBlurHashes.delete(image.id))
		}

		if (!backgroundThumbs.value[image.id] && !pendingThumbnails.has(image.id)) {
			pendingThumbnails.add(image.id)
			queryClient.ensureQueryData(unsplashBackgroundThumbnailQuery(image.id))
				.then(blob => {
					if (mounted && visibleImageIds.has(image.id)) {
						replaceObjectUrl(backgroundThumbs, image.id, blob)
					}
				})
				.catch(() => {})
				.finally(() => pendingThumbnails.delete(image.id))
		}
	})
}, {immediate: true})


async function setBackground(backgroundId: string) {
	if (setBackgroundMutation.isPending.value) {
		return
	}

	const updated = await setBackgroundMutation.mutateAsync({
		imageId: backgroundId,
		projectId: projectId.value,
	})
	await baseStore.handleSetCurrentProject({
		project: normalizeProject(updated),
		forceUpdate: true,
	})
	success({message: t('project.background.success')})
}

const backgroundUploadInput = ref<HTMLInputElement | null>(null)
async function uploadBackground() {
	const file = backgroundUploadInput.value?.files?.[0]
	if (!file) {
		return
	}

	const updated = await uploadBackgroundMutation.mutateAsync({projectId: projectId.value, file})
	await baseStore.handleSetCurrentProject({
		project: normalizeProject(updated),
		forceUpdate: true,
	})
	success({message: t('project.background.success')})
}

async function removeBackground() {
	const updated = await deleteBackgroundMutation.mutateAsync(projectId.value)
	await baseStore.handleSetCurrentProject({
		project: normalizeProject(updated),
		forceUpdate: true,
	})
	success({message: t('project.background.removeSuccess')})
	router.back()
}

onBeforeUnmount(() => {
	mounted = false
	visibleImageIds.clear()
	removeInvisibleObjectUrls(backgroundThumbs)
	removeInvisibleObjectUrls(backgroundBlurHashes)
})
</script>

<style lang="scss" scoped>
.unsplash-credit {
	text-align: end;
	font-size: .8rem;
}

.unsplash-credit__link {
	color: var(--grey-800);
}

.image-search__result-list {
	--items-per-row: 1;
	margin: 1rem 0 0;
	display: grid;
	gap: 1rem;
	grid-template-columns: repeat(var(--items-per-row), 1fr);

	@media screen and (min-width: $mobile) {
		--items-per-row: 2;
	}
	@media screen and (min-width: $tablet) {
		--items-per-row: 4;
	}
	@media screen and (min-width: $tablet) {
		--items-per-row: 5;
	}
}

.image-search__result-item {
	margin-block-start: 0; // FIXME: removes padding from .content
	aspect-ratio: 16 / 10;
	background-size: cover;
	background-position: center;
	display: flex;
	position: relative;
}

.image-search__image-button {
	inline-size: 100%;
}

.image-search__image {
	inline-size: 100%;
	block-size: 100%;
	object-fit: cover;
}

.image-search__info {
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
}
.image-search__result-item:hover .image-search__info {
		opacity: 1;
}

.is-load-more-button {
	margin: 1rem auto 0 !important;
	display: block;
	inline-size: 200px;
}
</style>
