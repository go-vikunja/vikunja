<template>
	<create-edit
		:title="$t('list.background.title')"
		primary-label=""
		:loading="backgroundService.loading"
		class="list-background-setting"
		:wide="true"
		v-if="uploadBackgroundEnabled || unsplashBackgroundEnabled"
		:tertary="hasBackground ? $t('list.background.remove') : ''"
		@tertary="removeBackground()"
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
				@click="$refs.backgroundUploadInput.click()"
				type="primary"
			>
				{{ $t('list.background.upload') }}
			</x-button>
		</div>
		<template v-if="unsplashBackgroundEnabled">
			<input
				:class="{'is-loading': backgroundService.loading}"
				@keyup="() => debounceNewBackgroundSearch()"
				class="input is-expanded"
				:placeholder="$t('list.background.searchPlaceholder')"
				type="text"
				v-model="backgroundSearchTerm"
			/>
			<p class="unsplash-link">
				<a href="https://unsplash.com" rel="noreferrer noopener nofollow" target="_blank">{{ $t('list.background.poweredByUnsplash') }}</a>
			</p>
			<div class="image-search-result">
				<a
					:key="im.id"
					:style="{'background-image': `url(${backgroundThumbs[im.id]})`}"
					@click="() => setBackground(im.id)"
					class="image"
					v-for="im in backgroundSearchResult">
					<a :href="`https://unsplash.com/@${im.info.author}`" rel="noreferrer noopener nofollow" target="_blank" class="info">
						{{ im.info.authorName }}
					</a>
				</a>
			</div>
			<x-button
				:disabled="backgroundService.loading"
				@click="() => searchBackgrounds(currentPage + 1)"
				class="is-load-more-button mt-4"
				:shadow="false"
				type="secondary"
				v-if="backgroundSearchResult.length > 0"
			>
				{{ backgroundService.loading ? $t('misc.loading') : $t('list.background.loadMore') }}
			</x-button>
		</template>
	</create-edit>
</template>

<script>
import {mapState} from 'vuex'
import BackgroundUnsplashService from '../../../services/backgroundUnsplash'
import BackgroundUploadService from '../../../services/backgroundUpload'
import ListService from '@/services/list'
import {CURRENT_LIST} from '@/store/mutation-types'
import CreateEdit from '@/components/misc/create-edit.vue'
import debounce from 'lodash.debounce'

const SEARCH_DEBOUNCE = 300

export default {
	name: 'list-setting-background',
	components: {CreateEdit},
	data() {
		return {
			backgroundService: new BackgroundUnsplashService(),
			backgroundSearchTerm: '',
			backgroundSearchResult: [],
			backgroundThumbs: {},
			currentPage: 1,

			// We're using debounce to not search on every keypress but with a delay.
			debounceNewBackgroundSearch: debounce(this.newBackgroundSearch, SEARCH_DEBOUNCE, {
				leading: true,
				trailing: true,
			}),

			backgroundUploadService: new BackgroundUploadService(),
			listService: new ListService(),
		}
	},
	computed: mapState({
		unsplashBackgroundEnabled: state => state.config.enabledBackgroundProviders.includes('unsplash'),
		uploadBackgroundEnabled: state => state.config.enabledBackgroundProviders.includes('upload'),
		currentList: state => state.currentList,
		hasBackground: state => state.background !== null,
	}),
	created() {
		this.setTitle(this.$t('list.background.title'))
		// Show the default collection of backgrounds
		this.newBackgroundSearch()
	},
	methods: {
		newBackgroundSearch() {
			if (!this.unsplashBackgroundEnabled) {
				return
			}
			// This is an extra method to reset a few things when searching to not break loading more photos.
			this.backgroundSearchResult = []
			this.backgroundThumbs = {}
			this.searchBackgrounds()
		},

		async searchBackgrounds(page = 1) {
			this.currentPage = page
			const result = await this.backgroundService.getAll({}, {s: this.backgroundSearchTerm, p: page})
			this.backgroundSearchResult = this.backgroundSearchResult.concat(result)
			result.forEach(async background => {
				this.backgroundThumbs[background.id] = await this.backgroundService.thumb(background)
			})
		},

		async setBackground(backgroundId) {
			// Don't set a background if we're in the process of setting one
			if (this.backgroundService.loading) {
				return
			}

			const list = await this.backgroundService.update({id: backgroundId, listId: this.$route.params.listId})
			await this.$store.dispatch(CURRENT_LIST, list)
			this.$store.commit('namespaces/setListInNamespaceById', list)
			this.$message.success({message: this.$t('list.background.success')})
		},

		async uploadBackground() {
			if (this.$refs.backgroundUploadInput.files.length === 0) {
				return
			}

			const list = await this.backgroundUploadService.create(this.$route.params.listId, this.$refs.backgroundUploadInput.files[0])
			await this.$store.dispatch(CURRENT_LIST, list)
			this.$store.commit('namespaces/setListInNamespaceById', list)
			this.$message.success({message: this.$t('list.background.success')})
		},

		async removeBackground() {
			const list = await this.listService.removeBackground(this.currentList)
			await this.$store.dispatch(CURRENT_LIST, list)
			this.$store.commit('namespaces/setListInNamespaceById', list)
			this.$message.success({message: this.$t('list.background.removeSuccess')})
			this.$router.back()
		},
	},
}
</script>
