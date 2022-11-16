<template>
	<create-edit
		v-if="uploadBackgroundEnabled || unsplashBackgroundEnabled"
		:title="$t('list.background.title')"
		:loading="backgroundService.loading"
		class="list-background-setting"
		:wide="true"
	>
		<div class="mb-4" v-if="uploadBackgroundEnabled">
			<input
				@change="uploadBackground"
				accept="image/*"
				class="is-hidden"
				ref="backgroundUploadInput"
				type="file"
			/>
			<x-button
				:loading="backgroundUploadService.loading"
				@click="backgroundUploadInput?.click()"
				variant="primary"
			>
				{{ $t('list.background.upload') }}
			</x-button>
		</div>
		<template v-if="unsplashBackgroundEnabled">
			<input
				:class="{'is-loading': backgroundService.loading}"
				@keyup="debounceNewBackgroundSearch()"
				class="input is-expanded"
				:placeholder="$t('list.background.searchPlaceholder')"
				type="text"
				v-model="backgroundSearchTerm"
			/>

			<p class="unsplash-credit">
				<BaseButton class="unsplash-credit__link" href="https://unsplash.com">{{ $t('list.background.poweredByUnsplash') }}</BaseButton>
			</p>

			<ul class="image-search__result-list">
				<li
					v-for="im in backgroundSearchResult"
					class="image-search__result-item"
					:key="im.id"
					:style="{'background-image': `url(${backgroundBlurHashes[im.id]})`}"
				>
					<CustomTransition name="fade">
						<BaseButton
							v-if="backgroundThumbs[im.id]"
							class="image-search__image-button"
							@click="setBackground(im.id)"
						>
							<img class="image-search__image" :src="backgroundThumbs[im.id]" alt="" />
						</BaseButton>
					</CustomTransition>

					<BaseButton
						:href="`https://unsplash.com/@${im.info.author}`"
						class="image-search__info"
					>
						{{ im.info.authorName }}
					</BaseButton>
				</li>
			</ul>
			<x-button
				v-if="backgroundSearchResult.length > 0"
				:disabled="backgroundService.loading"
				@click="searchBackgrounds(currentPage + 1)"
				class="is-load-more-button mt-4"
				:shadow="false"
				variant="secondary"
			>
				{{ backgroundService.loading ? $t('misc.loading') : $t('list.background.loadMore') }}
			</x-button>
		</template>

		<template #footer>
			<x-button
				v-if="hasBackground"
				:shadow="false"
				variant="tertiary"
				class="is-danger"
				@click.prevent.stop="removeBackground"
			>
				{{ $t('list.background.remove') }}
			</x-button>
			<x-button
				variant="secondary"
				@click.prevent.stop="$router.back()"
			>
				{{ $t('misc.close') }}
			</x-button>
		</template>
	</create-edit>
</template>

<script lang="ts">
export default { name: 'list-setting-background' }
</script>

<script setup lang="ts">
import {ref, computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'
import debounce from 'lodash.debounce'

import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import {useBaseStore} from '@/stores/base'
import {useListStore} from '@/stores/lists'
import {useNamespaceStore} from '@/stores/namespaces'
import {useConfigStore} from '@/stores/config'

import BackgroundUnsplashService from '@/services/backgroundUnsplash'
import BackgroundUploadService from '@/services/backgroundUpload'
import ListService from '@/services/list'
import type BackgroundImageModel from '@/models/backgroundImage'

import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'
import {useTitle} from '@/composables/useTitle'

import CreateEdit from '@/components/misc/create-edit.vue'
import {success} from '@/message'

const SEARCH_DEBOUNCE = 300

const {t} = useI18n({useScope: 'global'})
const baseStore = useBaseStore()
const route = useRoute()
const router = useRouter()

useTitle(() => t('list.background.title'))

const backgroundService = shallowReactive(new BackgroundUnsplashService())
const backgroundSearchTerm = ref('')
const backgroundSearchResult = ref([])
const backgroundThumbs = ref<Record<string, string>>({})
const backgroundBlurHashes = ref<Record<string, string>>({})
const currentPage = ref(1)

// We're using debounce to not search on every keypress but with a delay.
const debounceNewBackgroundSearch = debounce(newBackgroundSearch, SEARCH_DEBOUNCE, {
	trailing: true,
})

const backgroundUploadService = ref(new BackgroundUploadService())
const listService = ref(new ListService())
const listStore = useListStore()
const namespaceStore = useNamespaceStore()
const configStore = useConfigStore()

const unsplashBackgroundEnabled = computed(() => configStore.enabledBackgroundProviders.includes('unsplash'))
const uploadBackgroundEnabled = computed(() => configStore.enabledBackgroundProviders.includes('upload'))
const currentList = computed(() => baseStore.currentList)
const hasBackground = computed(() => baseStore.background !== null)

// Show the default collection of backgrounds
newBackgroundSearch()

function newBackgroundSearch() {
	if (!unsplashBackgroundEnabled.value) {
		return
	}
	// This is an extra method to reset a few things when searching to not break loading more photos.
	backgroundSearchResult.value = []
	backgroundThumbs.value = {}
	searchBackgrounds()
}

async function searchBackgrounds(page = 1) {
	currentPage.value = page
	const result = await backgroundService.getAll({}, {s: backgroundSearchTerm.value, p: page})
	backgroundSearchResult.value = backgroundSearchResult.value.concat(result)
	result.forEach((background: BackgroundImageModel) => {
		getBlobFromBlurHash(background.blurHash)
			.then((b) => {
				backgroundBlurHashes.value[background.id] = window.URL.createObjectURL(b)
			})

		backgroundService.thumb(background).then(b => {
			backgroundThumbs.value[background.id] = b
		})
	})
}


async function setBackground(backgroundId: string) {
	// Don't set a background if we're in the process of setting one
	if (backgroundService.loading) {
		return
	}

	const list = await backgroundService.update({
		id: backgroundId,
		listId: route.params.listId,
	})
	await baseStore.handleSetCurrentList({list, forceUpdate: true})
	namespaceStore.setListInNamespaceById(list)
	listStore.setList(list)
	success({message: t('list.background.success')})
}

const backgroundUploadInput = ref<HTMLInputElement | null>(null)
async function uploadBackground() {
	if (backgroundUploadInput.value?.files?.length === 0) {
		return
	}

	const list = await backgroundUploadService.value.create(
		route.params.listId,
		backgroundUploadInput.value?.files[0],
	)
	await baseStore.handleSetCurrentList({list, forceUpdate: true})
	namespaceStore.setListInNamespaceById(list)
	listStore.setList(list)
	success({message: t('list.background.success')})
}

async function removeBackground() {
	const list = await listService.value.removeBackground(currentList.value)
	await baseStore.handleSetCurrentList({list, forceUpdate: true})
	namespaceStore.setListInNamespaceById(list)
	listStore.setList(list)
	success({message: t('list.background.removeSuccess')})
	router.back()
}
</script>

<style lang="scss" scoped>
.unsplash-credit {
	text-align: right;
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
	margin-top: 0; // FIXME: removes padding from .content
	aspect-ratio: 16 / 10;
	background-size: cover;
	background-position: center;
	display: flex;
	position: relative;
}

.image-search__image-button {
	width: 100%;
}

.image-search__image {
	width: 100%;
	height: 100%;
	object-fit: cover;
}

.image-search__info {
	position: absolute;
	bottom: 0;
	width: 100%;
	padding: .25rem 0;
	opacity: 0;
	text-align: center;
	background: rgba(0, 0, 0, 0.5);
	font-size: .75rem;
	font-weight: bold;
	color: var(--white);
	transition: opacity $transition;
}
.image-search__result-item:hover .image-search__info {
		opacity: 1;
}

.is-load-more-button {
	margin: 1rem auto 0 !important;
	display: block;
	width: 200px;
}
</style>