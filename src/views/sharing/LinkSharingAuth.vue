<template>
	<div class="message is-centered is-info">
		<div class="message-header">
			<p class="has-text-centered">
				Authenticating...
			</p>
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
			hash: '',
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
			if (this.authLinkShare) {
				return
			}

			this.$store.dispatch('auth/linkShareAuth', this.$route.params.share)
				.then((r) => {
					console.log('after link share auth')
					this.$router.push({name: 'list.list', params: {listId: r.list_id}})
				})
				.catch(e => {
					this.error(e, this)
				})
		},
	},
}
</script>
