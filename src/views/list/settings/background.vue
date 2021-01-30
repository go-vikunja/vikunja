<template>
	<create-edit
		title="Set list background"
		primary-label=""
		:loading="backgroundService.loading"
		class="list-background-setting"
		:wide="true"
		v-if="uploadBackgroundEnabled || unsplashBackgroundEnabled"
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
				Choose a background from your pc
			</x-button>
		</div>
		<template v-if="unsplashBackgroundEnabled">
			<input
				:class="{'is-loading': backgroundService.loading}"
				@keyup="() => newBackgroundSearch()"
				class="input is-expanded"
				placeholder="Search for a background..."
				type="text"
				v-model="backgroundSearchTerm"
			/>
			<p class="unsplash-link"><a href="https://unsplash.com" target="_blank">Powered by Unsplash</a></p>
			<div class="image-search-result">
				<a
					:key="im.id"
					:style="{'background-image': `url(${backgroundThumbs[im.id]})`}"
					@click="() => setBackground(im.id)"
					class="image"
					v-for="im in backgroundSearchResult">
					<a :href="`https://unsplash.com/@${im.info.author}`" target="_blank" class="info">
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
				{{ backgroundService.loading ? 'Loading...' : 'Load more photos'}}
			</x-button>
		</template>
	</create-edit>
</template>

<script>
import BackgroundUnsplashService from '../../../services/backgroundUnsplash'
import BackgroundUploadService from '../../../services/backgroundUpload'
import {CURRENT_LIST} from '@/store/mutation-types'
import CreateEdit from '@/components/misc/create-edit'

export default {
	name: 'list-setting-background',
	components: {CreateEdit},
	data() {
		return {
			backgroundSearchTerm: '',
			backgroundSearchResult: [],
			backgroundService: null,
			backgroundThumbs: {},
			currentPage: 1,
			backgroundSearchTimeout: null,

			backgroundUploadService: null,
		}
	},
	computed: {
		unsplashBackgroundEnabled() {
			return this.$store.state.config.enabledBackgroundProviders.includes('unsplash')
		},
		uploadBackgroundEnabled() {
			return this.$store.state.config.enabledBackgroundProviders.includes('upload')
		},
	},
	created() {
		this.backgroundService = new BackgroundUnsplashService()
		this.backgroundUploadService = new BackgroundUploadService()
		this.setTitle('Set a list background')
		// Show the default collection of backgrounds
		this.newBackgroundSearch()
	},
	methods: {
		newBackgroundSearch() {
			if (!this.unsplashBackgroundEnabled) {
				return
			}
			// This is an extra method to reset a few things when searching to not break loading more photos.
			this.$set(this, 'backgroundSearchResult', [])
			this.$set(this, 'backgroundThumbs', {})
			this.searchBackgrounds()
		},
		searchBackgrounds(page = 1) {

			if (this.backgroundSearchTimeout !== null) {
				clearTimeout(this.backgroundSearchTimeout)
			}

			// We're using the timeout to not search on every keypress but with a 300ms delay.
			// If another key is pressed within these 300ms, the last search request is dropped and a new one is scheduled.
			this.backgroundSearchTimeout = setTimeout(() => {
				this.currentPage = page
				this.backgroundService.getAll({}, {s: this.backgroundSearchTerm, p: page})
					.then(r => {
						this.backgroundSearchResult = this.backgroundSearchResult.concat(r)
						r.forEach(b => {
							this.backgroundService.thumb(b)
								.then(t => {
									this.$set(this.backgroundThumbs, b.id, t)
								})
						})
					})
					.catch(e => {
						this.error(e, this)
					})
			}, 300)
		},
		setBackground(backgroundId) {
			// Don't set a background if we're in the process of setting one
			if (this.backgroundService.loading) {
				return
			}

			this.backgroundService.update({id: backgroundId, listId: this.$route.params.listId})
				.then(l => {
					this.$store.commit(CURRENT_LIST, l)
					this.$store.commit('namespaces/setListInNamespaceById', l)
					this.success({message: 'The background has been set successfully!'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
		uploadBackground() {
			if (this.$refs.backgroundUploadInput.files.length === 0) {
				return
			}

			this.backgroundUploadService.create(this.$route.params.listId, this.$refs.backgroundUploadInput.files[0])
				.then(l => {
					this.$store.commit(CURRENT_LIST, l)
					this.$store.commit('namespaces/setListInNamespaceById', l)
					this.success({message: 'The background has been set successfully!'}, this)
				})
				.catch(e => {
					this.error(e, this)
				})
		},
	},
}
</script>
