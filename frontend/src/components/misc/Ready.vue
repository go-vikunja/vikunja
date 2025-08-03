<template>
	<!-- This is a workaround to get the sw to "see" the to-be-cached version of the offline background image -->
	<div
		class="offline"
		style="height: 0;width: 0;"
	/>
	<div
		v-if="!online"
		class="app offline"
	>
		<div class="offline-message">
			<h1 class="title">
				{{ $t('offline.title') }}
			</h1>
			<p>{{ $t('offline.text') }}</p>
		</div>
	</div>
	<template v-else-if="baseStore.ready">
		<slot />
	</template>
	<section v-else-if="baseStore.error !== ''">
		<NoAuthWrapper>
			<p v-if="baseStore.error === ERROR_NO_API_URL">
				{{ $t('ready.noApiUrlConfigured') }}
			</p>
			<Message
				v-else
				variant="danger"
				class="mbe-4"
			>
				<p>
					{{ $t('ready.errorOccured') }}<br>
					{{ baseStore.error }}
				</p>
				<p>
					{{ $t('ready.checkApiUrl') }}
				</p>
			</Message>
			<ApiConfig
				:configure-open="true"
				@foundApi="baseStore.loadApp()"
			/>
		</NoAuthWrapper>
	</section>
	<CustomTransition name="fade">
		<section
			v-if="baseStore.loading"
			class="vikunja-loading"
		>
			<Logo class="logo" />
			<p>
				<span class="loader-container is-loading-small is-loading" />
				{{ $t('ready.loading') }}
			</p>
		</section>
	</CustomTransition>
</template>

<script lang="ts" setup>
import Logo from '@/assets/logo.svg?component'
import ApiConfig from '@/components/misc/ApiConfig.vue'
import Message from '@/components/misc/Message.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import NoAuthWrapper from '@/components/misc/NoAuthWrapper.vue'

import {ERROR_NO_API_URL} from '@/helpers/checkAndSetApiUrl'

import {useOnline} from '@/composables/useOnline'
import {useBaseStore} from '@/stores/base'

const online = useOnline()
const baseStore = useBaseStore()
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
	inset-inline-start: 0;
	bottom: 0;
	right: 0;
	background: var(--grey-100);
	z-index: 99;
}

.logo {
	margin-block-end: 1rem;
	width: 100px;
	height: 100px;
}

.loader-container {
	margin-inline-end: 1rem;

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
}

.title {
	font-weight: bold;
	font-size: 1.5rem;
	text-align: center;
	color: $white;
	font-weight: 700 !important;
	font-size: 1.5rem;
}
</style>
