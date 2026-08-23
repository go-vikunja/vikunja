<template>
	<CreateEdit
		v-if="uploadEnabled || unsplashEnabled"
		:title="$t(`${i18nPrefix}.title`)"
		:loading="unsplashService.loading"
		class="project-background-setting"
		:wide="true"
	>
		<div
			v-if="uploadEnabled"
			class="mbe-4"
		>
			<input
				ref="uploadInput"
				accept="image/*"
				class="is-hidden"
				type="file"
				@change="upload"
			>
			<XButton
				:loading="uploadService.loading"
				variant="primary"
				@click="uploadInput?.click()"
			>
				{{ $t(`${i18nPrefix}.upload`) }}
			</XButton>
		</div>
		<template v-if="unsplashEnabled">
			<input
				v-model="searchTerm"
				:class="{'is-loading': unsplashService.loading}"
				class="input is-expanded"
				:placeholder="$t(`${i18nPrefix}.searchPlaceholder`)"
				type="text"
				@keyup="debounceNewSearch()"
			>

			<p class="unsplash-credit">
				<BaseButton
					class="unsplash-credit__link"
					href="https://unsplash.com"
				>
					{{ $t(`${i18nPrefix}.poweredByUnsplash`) }}
				</BaseButton>
			</p>

			<ul
				class="image-search__result-list"
				:style="{'--image-aspect-ratio': aspectRatio}"
			>
				<li
					v-for="im in searchResult"
					:key="im.id"
					class="image-search__result-item"
					:style="{'background-image': `url(${blurHashes[im.id]})`}"
				>
					<CustomTransition name="fade">
						<BaseButton
							v-if="thumbs[im.id]"
							class="image-search__image-button"
							@click="set(im.id)"
						>
							<img
								class="image-search__image"
								:src="thumbs[im.id]"
								alt=""
							>
						</BaseButton>
					</CustomTransition>

					<BaseButton
						:href="`https://unsplash.com/@${im.info.author}?utm_source=vikunja&utm_medium=referral`"
						class="image-search__info"
					>
						{{ im.info.authorName }}
					</BaseButton>
				</li>
			</ul>
			<XButton
				v-if="searchResult.length > 0"
				:disabled="unsplashService.loading"
				class="is-load-more-button mbs-4"
				:shadow="false"
				variant="secondary"
				@click="search(currentPage + 1)"
			>
				{{ unsplashService.loading ? $t('misc.loading') : $t(`${i18nPrefix}.loadMore`) }}
			</XButton>
		</template>

		<template #footer>
			<XButton
				v-if="hasImage(currentProject)"
				:shadow="false"
				variant="tertiary"
				danger
				@click.prevent.stop="remove"
			>
				{{ $t(`${i18nPrefix}.remove`) }}
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
import {ref, computed} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'
import {useDebounceFn} from '@vueuse/core'

import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'

import type BackgroundImageModel from '@/models/backgroundImage'
import type {IProject} from '@/modelTypes/IProject'
import type {IFile} from '@/modelTypes/IFile'

import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'
import {useTitle} from '@/composables/useTitle'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import {success} from '@/message'

const props = withDefaults(defineProps<{
	i18nPrefix: string
	unsplashService: UnsplashServiceLike
	uploadService: UploadServiceLike
	uploadEnabled: boolean
	unsplashEnabled: boolean
	hasImage: (project: IProject) => boolean
	removeImage: (project: IProject) => Promise<IProject>
	aspectRatio?: string
}>(), {
	aspectRatio: '16 / 10',
})

defineOptions({name: 'ProjectImagePicker'})

// Duck-typed shapes shared by the background and card-image service pairs
// (backgroundUnsplash/cardBackgroundUnsplash, backgroundUpload/cardBackgroundUpload) —
// this component drives whichever pair is passed in, so it isn't tied to one image kind.
interface UnsplashServiceLike {
	loading: boolean
	getAll(model: object, params: {s: string, p: number}): Promise<BackgroundImageModel[]>
	update(model: {id: string, projectId: IProject['id']}): Promise<IProject>
	thumb(model: BackgroundImageModel): Promise<string>
}
interface UploadServiceLike {
	loading: boolean
	create(projectId: IProject['id'], file: IFile): Promise<IProject>
}

const SEARCH_DEBOUNCE = 300

const {t} = useI18n({useScope: 'global'})
const baseStore = useBaseStore()
const route = useRoute()
const router = useRouter()

useTitle(() => t(`${props.i18nPrefix}.title`))

const searchTerm = ref('')
const searchResult = ref<BackgroundImageModel[]>([])
const thumbs = ref<Record<string, string>>({})
const blurHashes = ref<Record<string, string>>({})
const currentPage = ref(1)

// We're using debounce to not search on every keypress but with a delay.
const debounceNewSearch = useDebounceFn(newSearch, SEARCH_DEBOUNCE)

const projectStore = useProjectStore()

const currentProject = computed(() => baseStore.currentProject)

// Show the default collection of images
newSearch()

function newSearch() {
	if (!props.unsplashEnabled) {
		return
	}
	// This is an extra method to reset a few things when searching to not break loading more photos.
	searchResult.value = []
	thumbs.value = {}
	search()
}

async function search(page = 1) {
	currentPage.value = page
	const result = await props.unsplashService.getAll({}, {s: searchTerm.value, p: page})
	searchResult.value = searchResult.value.concat(result)
	result.forEach((image: BackgroundImageModel) => {
		getBlobFromBlurHash(image.blurHash)
			.then((b) => {
				blurHashes.value[image.id] = window.URL.createObjectURL(b)
			})

		props.unsplashService.thumb(image).then(b => {
			thumbs.value[image.id] = b
		})
	})
}

async function set(imageId: string) {
	// Don't set an image if we're in the process of setting one
	if (props.unsplashService.loading) {
		return
	}

	const project = await props.unsplashService.update({
		id: imageId,
		projectId: route.params.projectId as unknown as IProject['id'],
	})
	await baseStore.handleSetCurrentProject({project, forceUpdate: true})
	projectStore.setProject(project)
	success({message: t(`${props.i18nPrefix}.success`)})
}

const uploadInput = ref<HTMLInputElement | null>(null)
async function upload() {
	if (uploadInput.value?.files?.length === 0) {
		return
	}

	const project = await props.uploadService.create(
		route.params.projectId as unknown as IProject['id'],
		uploadInput.value?.files[0] as unknown as IFile,
	)
	await baseStore.handleSetCurrentProject({project, forceUpdate: true})
	projectStore.setProject(project)
	success({message: t(`${props.i18nPrefix}.success`)})
}

async function remove() {
	const project = await props.removeImage(currentProject.value)
	await baseStore.handleSetCurrentProject({project, forceUpdate: true})
	projectStore.setProject(project)
	success({message: t(`${props.i18nPrefix}.removeSuccess`)})
	router.back()
}
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
	aspect-ratio: var(--image-aspect-ratio, 16 / 10);
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
