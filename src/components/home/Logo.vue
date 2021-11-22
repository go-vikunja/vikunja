<script setup lang="ts">
import { computed } from 'vue'
import LogoFull from '@/assets/logo-full.svg?component'
import LogoFullPride from '@/assets/logo-full-pride.svg?component'

const Logo = computed(() => new Date().getMonth() === 5 ? LogoFullPride : LogoFull)
</script>

<template>
	<Logo alt="Vikunja" class="logo" />
</template>

<style lang="scss" scoped>
.logo {
	color: var(--logo-text-color);
}
</style>