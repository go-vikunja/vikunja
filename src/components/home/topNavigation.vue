<template>
	<nav
		:class="{'has-background': background}"
		aria-label="main navigation"
		class="navbar main-theme is-fixed-top"
		role="navigation"
	>
		<div class="navbar-brand">
			<router-link :to="{name: 'home'}" class="navbar-item logo">
				<img alt="Vikunja" src="/images/logo-full-pride.svg" v-if="(new Date()).getMonth() === 5"/>
				<img alt="Vikunja" src="/images/logo-full.svg" v-else/>
			</router-link>
			<a
				@click="$store.commit('toggleMenu')"
				class="menu-show-button"
				@shortkey="() => $store.commit('toggleMenu')"
				v-shortkey="['ctrl', 'e']"
			>
				<icon icon="bars"></icon>
			</a>
		</div>
		<a
			@click="$store.commit('toggleMenu')"
			class="menu-show-button"
		>
			<icon icon="bars"></icon>
		</a>
		<div class="list-title" v-if="currentList.id">
			<h1
				:style="{ 'opacity': currentList.title === '' ? '0': '1' }"
				class="title">
				{{ currentList.title === '' ? 'Loading...' : currentList.title }}
			</h1>
			<router-link
				:to="{ name: 'list.edit', params: { id: currentList.id } }"
				class="icon"
				v-if="canWriteCurrentList">
				<icon icon="cog" size="2x"/>
			</router-link>
		</div>
		<div class="navbar-end">
			<update/>
			<div class="user">
				<img :src="userAvatar" alt="" class="avatar"/>
				<div class="dropdown is-right is-active">
					<div class="dropdown-trigger">
						<button @click.stop="userMenuActive = !userMenuActive" class="button noshadow">
							<span class="username">{{ userInfo.name !== '' ? userInfo.name : userInfo.username }}</span>
							<span class="icon is-small">
									<icon icon="chevron-down"/>
								</span>
						</button>
					</div>
					<transition name="fade">
						<div class="dropdown-menu" v-if="userMenuActive">
							<div class="dropdown-content">
								<router-link :to="{name: 'user.settings'}" class="dropdown-item">
									Settings
								</router-link>
								<a
									:href="imprintUrl"
									class="dropdown-item"
									target="_blank"
									v-if="imprintUrl">
									Imprint
								</a>
								<a
									:href="privacyPolicyUrl"
									class="dropdown-item"
									target="_blank"
									v-if="privacyPolicyUrl">
									Privacy policy
								</a>
								<a @click="$store.commit('keyboardShortcutsActive', true)" class="dropdown-item">Keyboard
									Shortcuts</a>
								<a @click="logout()" class="dropdown-item">
									Logout
								</a>
							</div>
						</div>
					</transition>
				</div>
			</div>
		</div>
	</nav>
</template>

<script>
import {mapState} from 'vuex'
import {CURRENT_LIST} from '@/store/mutation-types'
import Rights from '@/models/rights.json'
import Update from '@/components/home/update'

export default {
	name: 'topNavigation',
	data() {
		return {
			userMenuActive: false,
		}
	},
	components: {
		Update,
	},
	created() {
		// This will hide the menu once clicked outside of it
		this.$nextTick(() => document.addEventListener('click', () => this.userMenuActive = false))
	},
	computed: mapState({
		userInfo: state => state.auth.info,
		userAvatar: state => state.auth.avatarUrl,
		userAuthenticated: state => state.auth.authenticated,
		currentList: CURRENT_LIST,
		background: 'background',
		imprintUrl: state => state.config.legal.imprintUrl,
		privacyPolicyUrl: state => state.config.legal.privacyPolicyUrl,
		canWriteCurrentList: state => state.currentList.maxRight > Rights.READ,
	}),
	methods: {
		logout() {
			this.$store.dispatch('auth/logout')
			this.$router.push({name: 'user.login'})
		},
	},
}
</script>
