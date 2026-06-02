<template>
	<Card :title="$t('user.settings.feeds.title')">
		<p>
			{{ $t('user.settings.feeds.howTo') }}
		</p>
		<FormField
			v-model="feedUrl"
			type="text"
			readonly
		>
			<template #addon>
				<XButton
					v-tooltip="$t('misc.copy')"
					:shadow="false"
					icon="paste"
					@click="copy(feedUrl)"
				/>
			</template>
		</FormField>

		<p class="mbs-4">
			<i18n-t
				keypath="user.settings.feeds.usernameIs"
				scope="global"
			>
				<strong>{{ username }}</strong>
			</i18n-t>
		</p>

		<p class="mbs-2">
			<i18n-t
				keypath="user.settings.feeds.apiTokenHint"
				scope="global"
			>
				<template #scope>
					<code>feeds:access</code>
				</template>
				<template #link>
					<RouterLink
						:to="{
							name: 'user.settings.apiTokens',
							query: {
								title: $t('user.settings.feeds.tokenTitle'),
								scopes: 'feeds:access',
							},
						}"
					>
						{{ $t('user.settings.apiTokens.title') }}
					</RouterLink>
				</template>
			</i18n-t>
		</p>
	</Card>
</template>

<script lang="ts" setup>
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import {useTitle} from '@/composables/useTitle'
import {useCopyToClipboard} from '@/composables/useCopyToClipboard'
import FormField from '@/components/input/FormField.vue'
import {useConfigStore} from '@/stores/config'
import {useAuthStore} from '@/stores/auth'

const copy = useCopyToClipboard()

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.feeds.title')} - ${t('user.settings.title')}`)

const authStore = useAuthStore()
const configStore = useConfigStore()
const username = computed(() => authStore.info?.username)
const feedUrl = computed(() => `${configStore.apiBase}/feeds/notifications.atom`)
</script>
