<template>
	<div
			v-if="unsplashBackgroundEnabled"
			class="card list-background-setting loader-container"
			:class="{ 'is-loading': backgroundService.loading}">
		<header class="card-header">
			<p class="card-header-title">
				Set list background
			</p>
		</header>
		<div class="card-content">
			<div class="content">
				<input
						type="text"
						placeholder="Search for a background..."
						class="input is-expanded"
						v-model="backgroundSearchTerm"
						@keyup="() => newBackgroundSearch()"
						:class="{'is-loading': backgroundService.loading}"
				/>
				<div class="image-search-result">
					<a
							@click="() => setBackground(im.id)"
							class="image"
							v-for="im in backgroundSearchResult"
							:style="{'background-image': `url(${backgroundThumbs[im.id]})`}"
							:key="im.id">
						<a class="info" :href="`https://unsplash.com/@${im.info.author}`">
							{{ im.info.authorName }}
						</a>
					</a>
				</div>
				<a
						v-if="backgroundSearchResult.length > 0"
						class="button is-primary is-centered is-load-more-button is-outlined noshadow"
						@click="() => searchBackgrounds(currentPage + 1)"
						:disabled="backgroundService.loading"
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
	import {CURRENT_LIST} from '../../../store/mutation-types'

	export default {
		name: 'background',
		data() {
			return {
				backgroundSearchTerm: '',
				backgroundSearchResult: [],
				backgroundService: null,
				backgroundThumbs: {},
				currentPage: 1,
				backgroundSearchTimeout: null,
			}
		},
		props: {
			listId: {
				default: 0,
				required: true,
			}
		},
		computed: {
			unsplashBackgroundEnabled() {
				return this.$store.state.config.enabledBackgroundProviders.includes('unsplash')
			},
		},
		created() {
			this.backgroundService = new BackgroundUnsplashService()
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

				if(this.backgroundSearchTimeout !== null) {
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
						this.success({message: 'The background has been set successfully!'}, this)
					})
					.catch(e => {
						this.error(e, this)
					})
			},
		},
	}
</script>
