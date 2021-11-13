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
					<x-button @click="setApiUrl" :disabled="apiUrl === '' || null">
						{{ $t('apiConfig.change') }}
					</x-button>
				</div>
			</div>
		</div>
		<div class="api-url-info" v-else>
			<i18n-t keypath="apiConfig.signInOn">
				<span class="url" v-tooltip="apiUrl"> {{ apiDomain }} </span>
			</i18n-t>
			<br/>
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
import {parseURL} from 'ufo'
import {checkAndSetApiUrl} from '@/helpers/checkAndSetApiUrl'

export default {
	name: 'apiConfig',
	data() {
		return {
			configureApi: false,
			apiUrl: window.API_URL,
			errorMsg: '',
			successMsg: '',
		}
	},
	emits: ['foundApi'],
	created() {
		if (this.apiUrl === '') {
			this.configureApi = true
		}
	},
	computed: {
		apiDomain() {
			return parseURL(this.apiUrl).host
		},
	},
	props: {
		configureOpen: {
			type: Boolean,
			required: false,
			default: false,
		},
	},
	watch: {
		configureOpen: {
			handler(value) {
				this.configureApi = value
			},
			immediate: true,
		},
	},
	methods: {
		async setApiUrl() {
			if (this.apiUrl === '') {
				// Don't try to check and set an empty url
				this.errorMsg = this.$t('apiConfig.urlRequired')
				return
			}

			try {
				const url = await checkAndSetApiUrl(this.apiUrl)

				if (url === '') {
					// If the config setter function could not figure out a url					
					throw new Error('URL cannot be empty.')
				}

				// Set it + save it to local storage to save us the hoops
				this.errorMsg = ''
				this.successMsg = this.$t('apiConfig.success', {domain: this.apiDomain})
				this.configureApi = false
				this.apiUrl = url
				this.$emit('foundApi', this.apiUrl)
			} catch (e) {
				// Still not found, url is still invalid
				this.successMsg = ''
				this.errorMsg = this.$t('apiConfig.error', {domain: this.apiDomain})
			}
		},
	},
}
</script>

<style lang="scss" scoped>
.api-config {
	margin-bottom: .75rem;
}

.api-url-info {
	font-size: .9rem;
	text-align: right;
}

.url {
	border-bottom: 1px dashed $primary;
}
</style>