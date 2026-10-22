<template>
	<!-- The form seeds its draft once, so it may only mount with the account already loaded. -->
	<GeneralSettingsForm
		v-if="identity"
		:key="identity.id"
		:identity="identity"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import GeneralSettingsForm from './GeneralSettingsForm.vue'
import {useAuthStore} from '@/stores/auth'

defineOptions({name: 'UserSettingsGeneral'})

const authStore = useAuthStore()
const identity = computed(() => {
	const session = authStore.session
	if (!session || !authStore.currentAccount) return null
	return {
		id: session.id,
		type: session.type,
	}
})
</script>
