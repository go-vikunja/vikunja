<template>
    <a class="menu-bottom-link" :href="poweredByUrl" target="_blank" rel="noreferrer noopener nofollow">
        {{ $t('misc.poweredBy') }}
    </a>
</template>

<script setup lang="ts">
import {POWERED_BY as poweredByUrl} from '@/urls'
</script>

<style lang="scss">
.menu-bottom-link {
	color: $grey-300;
	text-align: center;
	display: block;
	padding-top: 1rem;
	padding-bottom: 1rem;
	font-size: .8rem;
}
</style>