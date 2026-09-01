<template>
	<CreateEdit
		v-if="uploadBackgroundEnabled || unsplashBackgroundEnabled"
		:title="$t('project.background.title')"
		:loading="backgroundMutationLoading"
		class="project-background-setting"
		:wide="true"
	>
		<div
			v-if="uploadBackgroundEnabled && canWrite"
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
		<template v-if="unsplashBackgroundEnabled && canWrite">
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
				>
					<UnsplashBackgroundThumbnail
						:image="im"
						@select="setBackground(im.id)"
					/>

					<BaseButton
						v-if="im.author !== ''"
						:href="`https://unsplash.com/@${encodeURIComponent(im.author)}?utm_source=vikunja&utm_medium=referral`"
						class="image-search__info"
					>
						{{ im.author_name }}
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
				v-if="hasBackground && canWrite"
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
import {computed, ref} from 'vue'
import {useInfiniteQuery, useQuery} from '@tanstack/vue-query'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'
import {useDebounceFn} from '@vueuse/core'

import {
	unsplashBackgroundSearchQuery,
	useDeleteProjectBackgroundMutation,
	useSetUnsplashProjectBackgroundMutation,
	useUploadProjectBackgroundMutation,
} from '@/client/queries/projectBackgrounds'
import {projectQuery} from '@/client/queries/projects'
import BaseButton from '@/components/base/BaseButton.vue'
import UnsplashBackgroundThumbnail from '@/components/project/partials/UnsplashBackgroundThumbnail.vue'

import {PERMISSIONS} from '@/constants/permissions'

import {useConfigStore} from '@/stores/config'
import {useBaseStore} from '@/stores/base'

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
const canWrite = computed(() => (project.data.value?.max_permission ?? 0) >= PERMISSIONS.READ_WRITE)
const backgroundSearchTerm = ref('')
const backgroundSearchInput = ref('')
const backgroundSearchQuery = useInfiniteQuery(computed(() => ({
	...unsplashBackgroundSearchQuery(backgroundSearchTerm.value),
	enabled: unsplashBackgroundEnabled.value,
})))
const backgroundSearchResult = computed(() => backgroundSearchQuery.data.value?.pages.flat() ?? [])

const hasBackground = computed(() => Boolean(project.data.value?.background_information))

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

async function setBackground(backgroundId: string) {
	if (setBackgroundMutation.isPending.value) {
		return
	}

	const id = projectId.value
	const updated = await setBackgroundMutation.mutateAsync({
		imageId: backgroundId,
		projectId: id,
	})
	// The user may have navigated to another project while the request was in flight
	if (baseStore.currentProjectId !== id) {
		return
	}
	await baseStore.handleSetCurrentProject({project: updated, forceUpdate: true})
	success({message: t('project.background.success')})
}

const backgroundUploadInput = ref<HTMLInputElement | null>(null)
async function uploadBackground() {
	const file = backgroundUploadInput.value?.files?.[0]
	if (!file) {
		return
	}

	const id = projectId.value
	const updated = await uploadBackgroundMutation.mutateAsync({projectId: id, file})
	if (baseStore.currentProjectId !== id) {
		return
	}
	await baseStore.handleSetCurrentProject({project: updated, forceUpdate: true})
	success({message: t('project.background.success')})
}

async function removeBackground() {
	const id = projectId.value
	const updated = await deleteBackgroundMutation.mutateAsync(id)
	if (baseStore.currentProjectId !== id) {
		return
	}
	await baseStore.handleSetCurrentProject({project: updated, forceUpdate: true})
	success({message: t('project.background.removeSuccess')})
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
	aspect-ratio: 16 / 10;
	background-size: cover;
	background-position: center;
	display: flex;
	position: relative;
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
