<template>
	<div class="api-config">
		<div v-if="configureApi">
			<label class="label" for="api-url">{{ $t('apiConfig.url') }}</label>
			<div class="field has-addons">
				<div class="control is-expanded">
					<input
						class="input"
						id="api-url"
						:placeholder="$t('apiConfig.urlPlaceholder')"
						required
						type="url"
						v-focus
						v-model="apiUrl"
						@keyup.enter="setApiUrl"
					/>
				</div>
				<div class="control">
					<x-button @click="setApiUrl" :disabled="apiUrl === ''">
						{{ $t('apiConfig.change') }}
					</x-button>
				</div>
			</div>
		</div>
		<div class="api-url-info" v-else>
			<i18n path="apiConfig.signInOn">
				<span class="url" v-tooltip="apiUrl"> {{ apiDomain() }} </span>
			</i18n>
			<br />
			<a @click="() => (configureApi = true)">{{ $t('apiConfig.change') }}</a>
		</div>

		<div
			class="notification is-success mt-2"
			v-if="successMsg !== '' && errorMsg === ''"
		>
			{{ successMsg }}
		</div>
		<div
			class="notification is-danger mt-2"
			v-if="errorMsg !== '' && successMsg === ''"
		>
			{{ errorMsg }}
		</div>
	</div>
</template>

<script>
export default {
	name: 'apiConfig',
	data() {
		return {
			configureApi: false,
			apiUrl: '',
			errorMsg: '',
			successMsg: '',
		}
	},
	created() {
		this.apiUrl = window.API_URL
		if (this.apiUrl === '') {
			this.configureApi = true
		}
	},
	methods: {
		apiDomain() {
			if (window.API_URL.startsWith('/api/v1')) {
				return window.location.host
			}
			const urlParts = window.API_URL.replace('http://', '')
				.replace('https://', '')
				.split(/[/?#]/)
			return urlParts[0]
		},
		setApiUrl() {
			if (this.apiUrl === '') {
				return
			}

			let urlToCheck = this.apiUrl

			// Check if the url has an http prefix
			if (
				!urlToCheck.startsWith('http://') &&
				!urlToCheck.startsWith('https://')
			) {
				urlToCheck = `http://${urlToCheck}`
			}

			urlToCheck = new URL(urlToCheck)
			const origUrlToCheck = urlToCheck

			const oldUrl = window.API_URL
			window.API_URL = urlToCheck.toString()

			// Check if the api is reachable at the provided url
			this.$store
				.dispatch('config/update')
				.catch((e) => {
					// Check if it is reachable at /api/v1 and http
					if (
						!urlToCheck.pathname.endsWith('/api/v1') &&
						!urlToCheck.pathname.endsWith('/api/v1/')
					) {
						urlToCheck.pathname = `${urlToCheck.pathname}api/v1`
						window.API_URL = urlToCheck.toString()
						return this.$store.dispatch('config/update')
					}
					return Promise.reject(e)
				})
				.catch((e) => {
					// Check if it has a port and if not check if it is reachable at https
					if (urlToCheck.protocol === 'http:') {
						urlToCheck.protocol = 'https:'
						window.API_URL = urlToCheck.toString()
						return this.$store.dispatch('config/update')
					}
					return Promise.reject(e)
				})
				.catch((e) => {
					// Check if it is reachable at /api/v1 and https
					urlToCheck.pathname = origUrlToCheck.pathname
					if (
						!urlToCheck.pathname.endsWith('/api/v1') &&
						!urlToCheck.pathname.endsWith('/api/v1/')
					) {
						urlToCheck.pathname = `${urlToCheck.pathname}api/v1`
						window.API_URL = urlToCheck.toString()
						return this.$store.dispatch('config/update')
					}
					return Promise.reject(e)
				})
				.catch((e) => {
					// Check if it is reachable at port 3456 and https
					if (urlToCheck.port !== 3456) {
						urlToCheck.protocol = 'https:'
						urlToCheck.port = 3456
						window.API_URL = urlToCheck.toString()
						return this.$store.dispatch('config/update')
					}
					return Promise.reject(e)
				})
				.catch((e) => {
					// Check if it is reachable at :3456 and /api/v1 and https
					urlToCheck.pathname = origUrlToCheck.pathname
					if (
						!urlToCheck.pathname.endsWith('/api/v1') &&
						!urlToCheck.pathname.endsWith('/api/v1/')
					) {
						urlToCheck.pathname = `${urlToCheck.pathname}api/v1`
						window.API_URL = urlToCheck.toString()
						return this.$store.dispatch('config/update')
					}
					return Promise.reject(e)
				})
				.catch((e) => {
					// Check if it is reachable at port 3456 and http
					if (urlToCheck.port !== 3456) {
						urlToCheck.protocol = 'http:'
						urlToCheck.port = 3456
						window.API_URL = urlToCheck.toString()
						return this.$store.dispatch('config/update')
					}
					return Promise.reject(e)
				})
				.catch((e) => {
					// Check if it is reachable at :3456 and /api/v1 and http
					urlToCheck.pathname = origUrlToCheck.pathname
					if (
						!urlToCheck.pathname.endsWith('/api/v1') &&
						!urlToCheck.pathname.endsWith('/api/v1/')
					) {
						urlToCheck.pathname = `${urlToCheck.pathname}api/v1`
						window.API_URL = urlToCheck.toString()
						return this.$store.dispatch('config/update')
					}
					return Promise.reject(e)
				})
				.catch(() => {
					// Still not found, url is still invalid
					this.successMsg = ''
					this.errorMsg = this.$t('apiConfig.error', {domain: this.apiDomain()})
					window.API_URL = oldUrl
				})
				.then((r) => {
					if (typeof r !== 'undefined') {
						// Set it + save it to local storage to save us the hoops
						this.errorMsg = ''
						this.successMsg = this.$t('apiConfig.success', {domain: this.apiDomain()})
						localStorage.setItem('API_URL', window.API_URL)
						this.configureApi = false
						this.apiUrl = window.API_URL
						this.$emit('foundApi', this.apiUrl)
					}
				})
		},
	},
}
</script>
