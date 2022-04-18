<template>
    <BaseButton class="menu-bottom-link" :href="poweredByUrl" target="_blank">
        {{ $t('misc.poweredBy') }}
    </BaseButton>
</template>

<script setup lang="ts">
import BaseButton from '@/components/base/BaseButton.vue'
import {POWERED_BY as poweredByUrl} from '@/urls'
</script>

<style lang="scss">
.menu-bottom-link {
	color: var(--grey-300);
	text-align: center;
	display: block;
	padding-top: 1rem;
	padding-bottom: 1rem;
	font-size: .8rem;
}
</style>