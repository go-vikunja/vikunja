<template>
	<div
		:class="{'has-background': background}"
		:style="{'background-image': `url(${background})`}"
		class="link-share-container"
	>
		<div class="container has-text-centered link-share-view">
			<div class="column is-10 is-offset-1">
				<img alt="Vikunja" class="logo" src="/images/logo-full.svg"/>
				<h1
					:style="{ 'opacity': currentList.title === '' ? '0': '1' }"
					class="title">
					{{ currentList.title === '' ? 'Loading...' : currentList.title }}
				</h1>
				<div class="box has-text-left view">
					<div class="logout">
						<x-button @click="logout()" type="secondary">
							<span>Logout</span>
							<span class="icon is-small">
								<icon icon="sign-out-alt"/>
							</span>
						</x-button>
					</div>
					<router-view/>
					<a class="menu-bottom-link" href="https://vikunja.io" target="_blank">
						Powered by Vikunja
					</a>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import {mapState} from 'vuex'
import {CURRENT_LIST} from '@/store/mutation-types'

export default {
	name: 'contentLinkShare',
	computed: mapState({
		currentList: CURRENT_LIST,
		background: 'background',
	}),
	methods: {
		logout() {
			this.$store.dispatch('auth/logout')
			this.$router.push({name: 'user.login'})
		},
	},
}
</script>
