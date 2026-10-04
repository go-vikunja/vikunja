<template>
	<div class="content">
		<RiskListPage :project-id="null" />
	</div>
</template>

<script setup lang="ts">
import {useI18n} from 'vue-i18n'

import RiskListPage from '@/components/risks/RiskListPage.vue'

import {useTitle} from '@/composables/useTitle'

defineOptions({name: 'RiskRegister'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('risks.register'))
</script>
