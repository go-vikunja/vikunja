<template>
	<div>
		<div class="notification is-info is-light has-text-centered" v-if="loading">
			Authenticating...
		</div>
		<div v-if="authenticateWithPassword" class="box">
			<p class="pb-2">
				This shared list requires a password. Please enter it below:
			</p>
			<div class="field">
				<div class="control">
					<input
						id="linkSharePassword"
						type="password"
						class="input"
						placeholder="e.g. ••••••••••••"
						v-model="password"
						v-focus
						@keyup.enter.prevent="auth"
					/>
				</div>
			</div>

			<x-button @click="auth" :loading="loading">
				Login
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
		this.setTitle('Authenticating...')
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

					let error = 'An error occured.'
					if (e.response && e.response.data && e.response.data.message) {
						error = e.response.data.message
					}
					if (typeof e.response.data.code !== 'undefined' && e.response.data.code === 13002) {
						error = 'The password is invalid.'
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
