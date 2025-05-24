<script setup lang="ts">
import { computed } from 'vue'
import { useNow } from '@vueuse/core'

import LogoFull from '@/assets/logo-full.svg?component'
import LogoFullPride from '@/assets/logo-full-pride.svg?component'
import {MILLISECONDS_A_HOUR} from '@/constants/date'

const now = useNow({
	interval: MILLISECONDS_A_HOUR,
})
const Logo = computed(() => window.ALLOW_ICON_CHANGES && now.value.getMonth() === 5 ? LogoFullPride : LogoFull)
const CustomLogo = computed(() => window.CUSTOM_LOGO_URL)
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
	max-width: 168px;
	max-height: 48px;
}
</style>
