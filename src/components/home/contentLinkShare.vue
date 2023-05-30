<template>
	<div
		:class="[background ? 'has-background' : '', $route.name as string +'-view']"
		:style="{'background-image': `url(${background})`}"
		class="link-share-container"
	>
		<div class="container has-text-centered link-share-view">
			<div class="column is-10 is-offset-1">
				<Logo class="logo" v-if="logoVisible"/>
				<h1
					:class="{'m-0': !logoVisible}"
					:style="{ 'opacity': currentProject?.title === '' ? '0': '1' }"
					class="title">
					{{ currentProject?.title === '' ? $t('misc.loading') : currentProject?.title }}
				</h1>
				<div class="box has-text-left view">
					<router-view/>
					<PoweredByLink/>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {computed} from 'vue'

import {useBaseStore} from '@/stores/base'

import Logo from '@/components/home/Logo.vue'
import PoweredByLink from './PoweredByLink.vue'

const baseStore = useBaseStore()
const currentProject = computed(() => baseStore.currentProject)
const background = computed(() => baseStore.background)
const logoVisible = computed(() => baseStore.logoVisible)
</script>

<style lang="scss" scoped>
.link-share-container.has-background .view {
	background-color: transparent;
	border: none;
}

.logo {
	max-width: 300px;
	width: 90%;
	margin: 2rem 0 1.5rem;
	height: 100px;
}

.column {
	max-width: 100%;
}

.title {
	text-shadow: 0 0 1rem var(--white);
}

// FIXME: this should be defined somewhere deep
.link-share-view .card {
	background-color: var(--white);
}
</style>
