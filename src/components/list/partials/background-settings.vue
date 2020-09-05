<template>
	<div
		:class="{ 'is-loading': backgroundService.loading}"
		class="card list-background-setting loader-container"
		v-if="uploadBackgroundEnabled || unsplashBackgroundEnabled">
		<header class="card-header">
			<p class="card-header-title">
				Set list background
			</p>
		</header>
		<div class="card-content">
			<div class="content" v-if="uploadBackgroundEnabled">
				<input
					@change="uploadBackground"
					accept="image/*"
					class="is-hidden"
					ref="backgroundUploadInput"
					type="file"
				/>
				<a
					:class="{'is-loading': backgroundUploadService.loading}"
					@click="$refs.backgroundUploadInput.click()"
					class="button is-primary"
				>
					Choose a background from your pc
				</a>
			</div>
			<div class="content" v-if="unsplashBackgroundEnabled">
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
						<a :href="`https://unsplash.com/@${im.info.author}`" class="info">
							{{ im.info.authorName }}
						</a>
					</a>
				</div>
				<a
					:disabled="backgroundService.loading"
					@click="() => searchBackgrounds(currentPage + 1)"
					class="button is-primary is-centered is-load-more-button is-outlined noshadow"
					v-if="backgroundSearchResult.length > 0"
				>
					<template v-if="backgroundService.loading">
						Loading...
					</template>
					<template v-else>
						Load more photos
					</template>
				</a>
			</div>
		</div>
	</div>
</template>

<script>
import BackgroundUnsplashService from '../../../services/backgroundUnsplash'
import BackgroundUploadService from '../../../services/backgroundUpload'
import {CURRENT_LIST} from '@/store/mutation-types'

export default {
	name: 'background-settings',
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
	props: {
		listId: {
			default: 0,
			required: true,
		},
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

			this.backgroundService.update({id: backgroundId, listId: this.listId})
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

			this.backgroundUploadService.create(this.listId, this.$refs.backgroundUploadInput.files[0])
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
