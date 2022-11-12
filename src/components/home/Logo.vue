<script setup lang="ts">
import {computed} from 'vue'
import {useNow} from '@vueuse/core'

import LogoFull from '@/assets/logo-full.svg?url'
import LogoFullPride from '@/assets/logo-full-pride.svg?url'

const now = useNow()
const logoUrl = computed(() => now.value.getMonth() === 5 ? LogoFullPride : LogoFull)
</script>

<template>
	<img alt="Vikunja" :src="logoUrl" class="logo" />
</template>

<style lang="scss" scoped>
.logo {
	color: var(--logo-text-color);
}
</style>