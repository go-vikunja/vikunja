<script setup lang="ts">
import { computed } from 'vue'
import { useNow } from '@vueuse/core'
import { useAuthStore } from '@/stores/auth'
import { useColorScheme } from '@/composables/useColorScheme'

import LogoFull from '@/assets/logo-full.svg?component'
import LogoFullPride from '@/assets/logo-full-pride.svg?component'
import {MILLISECONDS_A_HOUR} from '@/constants/date'

const now = useNow({
	interval: MILLISECONDS_A_HOUR,
})

const authStore = useAuthStore()
const { isDark } = useColorScheme()

const Logo = computed(() => window.ALLOW_ICON_CHANGES
	&& authStore.settings.frontendSettings.allowIconChanges
	&& now.value.getMonth() === 5
	? LogoFullPride
	: LogoFull)

const CustomLogo = computed(() => {
	const lightLogo = window.CUSTOM_LOGO_URL
	const darkLogo = window.CUSTOM_LOGO_URL_DARK

	if (!lightLogo && !darkLogo) return ''
	if (!darkLogo) return lightLogo
	if (!lightLogo) return darkLogo

	return isDark.value ? darkLogo : lightLogo
})
</script>

<template>
	<div>
		<Logo
			v-if="!CustomLogo"
			alt="Vikunja"
			class="logo"
		/>
		<img
			v-show="CustomLogo"
			:src="CustomLogo"
			alt="Vikunja"
			class="logo"
		>
	</div>
</template>

<style lang="scss" scoped>
.logo {
	color: var(--logo-text-color);
	max-inline-size: 168px;
	max-block-size: 48px;
}
</style>
