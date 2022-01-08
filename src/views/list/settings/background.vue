<template>
	<create-edit
		:title="$t('list.background.title')"
		primary-label=""
		:loading="backgroundService.loading"
		class="list-background-setting"
		:wide="true"
		v-if="uploadBackgroundEnabled || unsplashBackgroundEnabled"
		:tertiary="hasBackground ? $t('list.background.remove') : ''"
		@tertiary="removeBackground()"
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
				variant="primary"
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
				variant="secondary"
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

<style lang="scss" scoped>
.list-background-setting {

  .unsplash-link {
    text-align: right;
    font-size: .8rem;

    a {
      color: var(--grey-800);
    }
  }

  .image-search-result {
    margin-top: 1rem;
    display: flex;
    flex-flow: row wrap;

    .image {
      width: calc(100% / 5 - 1rem);
      height: 120px;
      margin: .5rem;
      background-size: cover;
      background-position: center;
      display: flex;

      @media screen and (min-width: $desktop) {
        &:nth-child(5n) {
          break-after: always;
        }
      }

      @media screen and (max-width: $desktop) {
        width: calc(100% / 4 - 1rem);

        &:nth-child(4n) {
          break-after: always;
        }
      }

      @media screen and (max-width: $tablet) {
        width: calc(100% / 2 - 1rem);

        &:nth-child(2n) {
          break-after: always;
        }
      }

      @media screen and (max-width: ($mobile)) {
        width: calc(100% - 1rem);

        &:nth-child(1n) {
          break-after: always;
        }
      }

      .info {
        align-self: flex-end;
        display: block;
        opacity: 0;
        width: 100%;
        padding: .25rem 0;
        text-align: center;
        background: rgba(0, 0, 0, 0.5);
        font-size: .75rem;
        font-weight: bold;
        color: var(--white);
        transition: opacity $transition;
      }

      &:hover .info {
        opacity: 1;
      }
    }
  }

  .is-load-more-button {
    margin: 1rem auto 0 !important;
    display: block;
    width: 200px;
  }
}
</style>