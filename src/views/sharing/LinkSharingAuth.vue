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

			<div class="notification is-danger mt-4" v-if="errorMessage !== ''">
				{{ errorMessage }}
			</div>
		</div>
	</div>
</template>

<script>
import {mapGetters} from 'vuex'

export default {
	name: 'LinkSharingAuth',
	data() {
		return {
			loading: true,
			authenticateWithPassword: false,
			errorMessage: '',

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
	computed: mapGetters('auth', [
		'authLinkShare',
	]),
	methods: {
		async auth() {
			this.errorMessage = ''

			if (this.authLinkShare) {
				return
			}

			this.loading = true

			try {
				const r = await this.$store.dispatch('auth/linkShareAuth', {
					hash: this.$route.params.share,
					password: this.password,
				})
				this.$router.push({name: 'list.list', params: {listId: r.list_id}})
			} catch(e) {
				if (typeof e.response.data.code !== 'undefined' && e.response.data.code === 13001) {
					this.authenticateWithPassword = true
					return
				}

				// TODO: Put this logic in a global errorMessage handler method which checks all auth codes
				let errorMessage = this.$t('sharing.error')
				if (e.response && e.response.data && e.response.data.message) {
					errorMessage = e.response.data.message
				}
				if (typeof e.response.data.code !== 'undefined' && e.response.data.code === 13002) {
					errorMessage = this.$t('sharing.invalidPassword')
				}
				this.errorMessage = errorMessage
			} finally {
				this.loading = false
			}
		},
	},
}
</script>
