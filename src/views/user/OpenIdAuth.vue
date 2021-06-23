<template>
	<div>
		<div class="notification is-danger" v-if="errorMessage">
			{{ errorMessage }}
		</div>
		<div class="notification is-info" v-if="loading">
			{{ $t('user.auth.authenticating') }}
		</div>
	</div>
</template>

<script>
import {mapState} from 'vuex'

import {ERROR_MESSAGE, LOADING} from '@/store/mutation-types'
import {getErrorText} from '@/message'

export default {
	name: 'Auth',
	computed: mapState({
		errorMessage: ERROR_MESSAGE,
		loading: LOADING,
	}),
	mounted() {
		this.authenticateWithCode()
	},
	methods: {
		authenticateWithCode() {
			// This component gets mounted twice: The first time when the actual auth request hits the frontend,
			// the second time after that auth request succeeded and the outer component "content-no-auth" isn't used
			// but instead the "content-auth" component is used. Because this component is just a route and thus
			// gets mounted as part of a <router-view/> which both the content-auth and content-no-auth components have,
			// this re-mounts the component, even if the user is already authenticated.
			// To make sure we only try to authenticate the user once, we set this "authenticating" lock in localStorage
			// which ensures only one auth request is done at a time. We don't simply check if the user is already
			// authenticated to not prevent the whole authentication if some user is already logged in.
			if (localStorage.getItem('authenticating')) {
				return
			}
			localStorage.setItem('authenticating', true)

			const state = localStorage.getItem('state')
			if(typeof this.$route.query.state === 'undefined' || this.$route.query.state !== state) {
				localStorage.removeItem('authenticating')
				this.$store.commit(ERROR_MESSAGE, this.$t('user.auth.openIdStateError'))
				return
			}

			this.$store.commit(ERROR_MESSAGE, '')

			this.$store.dispatch('auth/openIdAuth', {
				provider: this.$route.params.provider,
				code: this.$route.query.code,
			})
				.then(() => {
					this.$router.push({name: 'home'})
				})
				.catch(e => {
					const err = getErrorText(e, p => this.$t(p))
					if (typeof err[1] !== 'undefined') {
						this.$store.commit(ERROR_MESSAGE, err[1])
						return
					}

					this.$store.commit(ERROR_MESSAGE, err[0])
				})
				.finally(() => {
					localStorage.removeItem('authenticating')
				})
		},
	},
}
</script>
