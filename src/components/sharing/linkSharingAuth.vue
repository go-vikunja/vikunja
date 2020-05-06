<template>
	<div class="message is-centered is-info" v-if="loading">
		<div class="message-header">
			<p class="has-text-centered">
				Authenticating...
			</p>
		</div>
	</div>
</template>

<script>
	import auth from '../../auth'
	import router from '../../router'

	export default {
		name: 'linkSharingAuth',
		data() {
			return {
				hash: '',
				loading: true,
			}
		},
		created() {
			this.auth()
		},
		methods: {
			auth() {
				auth.linkShareAuth(this.$route.params.share)
					.then((r) => {
						this.loading = false
						router.push({name: 'list.list', params: {listId: r.list_id}})
					})
					.catch(e => {
						this.error(e, this)
					})
			}
		},
	}
</script>
