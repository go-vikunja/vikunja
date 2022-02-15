<template>
	<div>
		<message variant="danger" v-if="errorMessage">
			{{ errorMessage }}
		</message>
		<message v-if="loading">
			{{ $t('user.auth.authenticating') }}
		</message>
	</div>
</template>

<script lang="ts">
import {mapState} from 'vuex'

import {LOADING} from '@/store/mutation-types'
import {getErrorText} from '@/message'
import Message from '@/components/misc/message'
import {clearLastVisited, getLastVisited} from '../../helpers/saveLastVisited'

export default {
	name: 'Auth',
	components: {Message},
	data() {
		return {
			errorMessage: '',
		}
	},
	computed: mapState({
		loading: LOADING,
	}),
	mounted() {
		this.authenticateWithCode()
	},
	methods: {
		async authenticateWithCode() {
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

			this.errorMessage = ''

			if (typeof this.$route.query.error !== 'undefined') {
				localStorage.removeItem('authenticating')
				this.errorMessage = typeof this.$route.query.message !== 'undefined'
					? this.$route.query.message
					: this.$t('user.auth.openIdGeneralError')
				return
			}

			const state = localStorage.getItem('state')
			if (typeof this.$route.query.state === 'undefined' || this.$route.query.state !== state) {
				localStorage.removeItem('authenticating')
				this.errorMessage = this.$t('user.auth.openIdStateError')
				return
			}

			try {
				await this.$store.dispatch('auth/openIdAuth', {
					provider: this.$route.params.provider,
					code: this.$route.query.code,
				})
				const last = getLastVisited()
				if (last !== null) {
					this.$router.push({
						name: last.name,
						params: last.params,
					})
					clearLastVisited()
				} else {
					this.$router.push({name: 'home'})
				}
			} catch(e) {
				const err = getErrorText(e)
				this.errorMessage = typeof err[1] !== 'undefined' ? err[1] : err[0]
			} finally {
				localStorage.removeItem('authenticating')
			}
		},
	},
}
</script>
