<script lang="ts" setup>
import {useI18n} from 'vue-i18n'

import Card from '@/components/misc/Card.vue'

import {useTitle} from '@/composables/useTitle'

defineOptions({name: 'UserSettingsWebhooks'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.webhooks.title')} - ${t('user.settings.title')}`)
</script>

<template>
	<Card
		:title="$t('user.settings.webhooks.title')"
	>
		<p class="mb-4">
			{{ $t('user.settings.webhooks.description') }}
		</p>
	</Card>
</template>
