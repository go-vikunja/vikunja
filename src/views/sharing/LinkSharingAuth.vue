<template>
	<div>
		<div class="notification is-info is-light has-text-centered" v-if="loading">
			{{ $t('sharing.authenticating') }}
		</div>
		<div v-if="authenticateWithPassword" class="box">
			<p class="pb-2">
				{{ $t('sharing.passwordRequired') }}
			</p>
			<div class="field">
				<div class="control">
					<input
						id="linkSharePassword"
						type="password"
						class="input"
						:placeholder="$t('user.auth.passwordPlaceholder')"
						v-model="password"
						v-focus
						@keyup.enter.prevent="auth"
					/>
				</div>
			</div>

			<x-button @click="auth" :loading="loading">
				{{ $t('user.auth.login') }}
			</x-button>

			<div class="notification is-danger mt-4" v-if="error !== ''">
				{{ error }}
			</div>
		</div>
	</div>
</template>

<script>
import {mapState} from 'vuex'
import authTypes from '@/models/authTypes.json'

export default {
	name: 'LinkSharingAuth',
	data() {
		return {
			loading: true,
			authenticateWithPassword: false,
			error: '',

			hash: '',
			password: '',
		}
	},
	created() {
		this.auth()
	},
	mounted() {
		this.setTitle(this.$t('sharing.authenticating'))
	},
	computed: mapState({
		authLinkShare: state => state.auth.authenticated && (state.auth.info && state.auth.info.type === authTypes.LINK_SHARE),
	}),
	methods: {
		auth() {
			this.error = ''

			if (this.authLinkShare) {
				return
			}

			this.loading = true

			this.$store.dispatch('auth/linkShareAuth', {hash: this.$route.params.share, password: this.password})
				.then((r) => {
					this.$router.push({name: 'list.list', params: {listId: r.list_id}})
				})
				.catch(e => {
					if (typeof e.response.data.code !== 'undefined' && e.response.data.code === 13001) {
						this.authenticateWithPassword = true
						return
					}

					// TODO: Put this logic in a global error handler method which checks all auth codes
					let error = this.$t('sharing.error')
					if (e.response && e.response.data && e.response.data.message) {
						error = e.response.data.message
					}
					if (typeof e.response.data.code !== 'undefined' && e.response.data.code === 13002) {
						error = this.$t('sharing.invalidPassword')
					}
					this.error = error
				})
				.finally(() => {
					this.loading = false
				})
		},
	},
}
</script>
