<template>
	<!-- This is a workaround to get the sw to "see" the to-be-cached version of the offline background image -->
	<div class="offline" style="height: 0;width: 0;"></div>
	<div class="app offline" v-if="!online">
		<div class="offline-message">
			<h1>{{ $t('offline.title') }}</h1>
			<p>{{ $t('offline.text') }}</p>
		</div>
	</div>
	<template v-else-if="ready">
		<slot/>
	</template>
	<section v-else-if="error !== ''">
		<no-auth-wrapper>
			<card>
				<p v-if="error === ERROR_NO_API_URL">
					{{ $t('ready.noApiUrlConfigured') }}
				</p>
				<message variant="danger" v-else>
					<p>
						{{ $t('ready.errorOccured') }}<br/>
						{{ error }}
					</p>
					<p>
						{{ $t('ready.checkApiUrl') }}
					</p>
				</message>
				<api-config :configure-open="true" @found-api="load"/>
			</card>
		</no-auth-wrapper>
	</section>
	<transition name="fade">
		<section class="vikunja-loading" v-if="showLoading">
			<Logo class="logo"/>
			<p>
				<span class="loader-container is-loading-small is-loading"></span>
				{{ $t('ready.loading') }}
			</p>
		</section>
	</transition>
</template>

<script lang="ts" setup>
import {ref, computed} from 'vue'
import {useStore} from 'vuex'

import Logo from '@/assets/logo.svg?component'
import ApiConfig from '@/components/misc/api-config.vue'
import Message from '@/components/misc/message.vue'
import NoAuthWrapper from '@/components/misc/no-auth-wrapper.vue'

import {ERROR_NO_API_URL} from '@/helpers/checkAndSetApiUrl'
import {useOnline} from '@/composables/useOnline'

const store = useStore()

const ready = computed(() => store.state.vikunjaReady)
const online = useOnline()

const error = ref('')
const showLoading = computed(() => !ready.value && error.value === '')

async function load() {
	try {
		await store.dispatch('loadApp')
	} catch(e: any) {
		error.value = e
	}
}

load()
</script>

<style lang="scss" scoped>
.vikunja-loading {
	display: flex;
	justify-content: center;
	align-items: center;
	height: 100vh;
	width: 100vw;
	flex-direction: column;
	position: fixed;
	top: 0;
	left: 0;
	bottom: 0;
	right: 0;
	background: var(--grey-100);
	z-index: 99;
}

.logo {
	margin-bottom: 1rem;
	width: 100px;
	height: 100px;
}

.loader-container {
	margin-right: 1rem;

	&.is-loading::after {
		border-left-color: var(--grey-400);
		border-bottom-color: var(--grey-400);
	}
}

.offline {
	background: url('@/assets/llama-nightscape.jpg') no-repeat center;
	background-size: cover;
	height: 100vh;
}

.offline-message {
	text-align: center;
	position: absolute;
	width: 100vw;
	bottom: 5vh;
	color: $white;
	padding: 0 1rem;

	h1 {
		font-weight: bold;
		font-size: 1.5rem;
		text-align: center;
		color: $white;
		font-weight: 700 !important;
		font-size: 1.5rem;
	}
}
</style>
