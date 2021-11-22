<template>
	<div
		:class="[background ? 'has-background' : '', $route.name+'-view']"
		:style="{'background-image': `url(${background})`}"
		class="link-share-container"
	>
		<div class="container has-text-centered link-share-view">
			<div class="column is-10 is-offset-1">
				<Logo class="logo" />
				<h1
					:style="{ 'opacity': currentList.title === '' ? '0': '1' }"
					class="title">
					{{ currentList.title === '' ? $t('misc.loading') : currentList.title }}
				</h1>
				<div class="box has-text-left view">
					<router-view/>
					<PoweredByLink />
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import {mapState} from 'vuex'

import Logo from '@/components/home/Logo.vue'
import PoweredByLink from './PoweredByLink.vue'

export default {
	name: 'contentLinkShare',
	components: {
		Logo,
		PoweredByLink,
	},
	computed: mapState([
		'currentList',
		'background',
	]),
}
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
