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
			<i18n-t keypath="apiConfig.use">
				<span class="url" v-tooltip="apiUrl"> {{ apiDomain }} </span>
			</i18n-t>
			<br/>
			<a @click="() => (configureApi = true)">{{ $t('apiConfig.change') }}</a>
		</div>

		<message variant="success" v-if="successMsg !== '' && errorMsg === ''" class="mt-2">
			{{ successMsg }}
		</message>
		<message variant="danger" v-if="errorMsg !== '' && successMsg === ''" class="mt-2">
			{{ errorMsg }}
		</message>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {parseURL} from 'ufo'

import {checkAndSetApiUrl} from '@/helpers/checkAndSetApiUrl'
import {success} from '@/message'

import Message from '@/components/misc/message.vue'

const props = defineProps({
	configureOpen: {
		type: Boolean,
		required: false,
		default: false,
	},
})
const emit = defineEmits(['foundApi'])

const apiUrl = ref(window.API_URL)
const configureApi = ref(apiUrl.value === '')

// Because we're only using this to parse the hostname, it should be fine to just prefix with http:// 
// regardless of whether the url is actually reachable under http.
const apiDomain = computed(() => parseURL(apiUrl.value, 'http://').host || parseURL(window.location.href).host)

watch(() => props.configureOpen, (value) => {
	configureApi.value = value
}, {immediate: true})


const {t} = useI18n()

const errorMsg = ref('')
const successMsg = ref('')

async function setApiUrl() {
	if (apiUrl.value === '') {
		// Don't try to check and set an empty url
		errorMsg.value = t('apiConfig.urlRequired')
		return
	}

	try {
		const url = await checkAndSetApiUrl(apiUrl.value)

		if (url === '') {
			// If the config setter function could not figure out a url					
			throw new Error('URL cannot be empty.')
		}

		// Set it + save it to local storage to save us the hoops
		errorMsg.value = ''
		apiUrl.value = url
		success({message: t('apiConfig.success', {domain: apiDomain.value})})
		configureApi.value = false
		emit('foundApi', apiUrl.value)
	} catch (e) {
		// Still not found, url is still invalid
		successMsg.value = ''
		errorMsg.value = t('apiConfig.error', {domain: apiDomain.value})
	}
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
	border-bottom: 1px dashed var(--primary);
}
</style>