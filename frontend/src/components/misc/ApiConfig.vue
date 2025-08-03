<template>
	<div class="api-config">
		<div v-if="configureApi">
			<label
				class="label"
				for="api-url"
			>{{ $t('apiConfig.url') }}</label>
			<div class="field has-addons">
				<div class="control is-expanded">
					<input
						id="api-url"
						v-model="apiUrl"
						v-focus
						class="input"
						:placeholder="$t('apiConfig.urlPlaceholder')"
						required
						type="url"
						@keyup.enter="setApiUrl"
					>
				</div>
				<div class="control">
					<XButton
						:disabled="apiUrl === '' || undefined"
						@click="setApiUrl"
					>
						{{ $t('apiConfig.change') }}
					</XButton>
				</div>
			</div>
		</div>
		<div
			v-else
			class="api-url-info"
		>
			<i18n-t
				keypath="apiConfig.use"
				scope="global"
			>
				<span
					v-tooltip="apiUrl"
					class="url"
				> {{ apiDomain }} </span>
			</i18n-t>
			<br>
			<ButtonLink
				class="api-config__change-button"
				@click="() => (configureApi = true)"
			>
				{{ $t('apiConfig.change') }}
			</ButtonLink>
		</div>

		<Message
			v-if="errorMsg !== ''"
			variant="danger"
			class="mbs-2"
		>
			{{ errorMsg }}
		</Message>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {parseURL} from 'ufo'

import {checkAndSetApiUrl} from '@/helpers/checkAndSetApiUrl'
import {success} from '@/message'

import Message from '@/components/misc/Message.vue'
import ButtonLink from '@/components/misc/ButtonLink.vue'

const props = withDefaults(defineProps<{
	configureOpen?: boolean
}>(), {
	configureOpen: false,
})
const emit = defineEmits<{
	'foundApi': [url: string],
}>()

const apiUrl = ref(window.API_URL)
const configureApi = ref(window.API_URL === '')

// Because we're only using this to parse the hostname, it should be fine to just prefix with http:// 
// regardless of whether the url is actually reachable under http.
const apiDomain = computed(() => parseURL(apiUrl.value, 'http://').host || parseURL(window.location.href).host)

watch(() => props.configureOpen, (value) => {
	configureApi.value = value
}, {immediate: true})


const {t} = useI18n({useScope: 'global'})

const errorMsg = ref('')

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
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
	} catch (e) {
		// Still not found, url is still invalid
		errorMsg.value = t('apiConfig.error', {domain: apiDomain.value})
	}
}
</script>

<style lang="scss" scoped>
.api-config {
	margin-block-end: .75rem;
}

.api-url-info {
	font-size: .9rem;
	text-align: end;
}

.url {
	border-inline-end: 1px dashed var(--primary);
}

.api-config__change-button {
	display: inline-block;
    color: var(--link);
}
</style>
