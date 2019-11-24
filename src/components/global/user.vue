<template>
	<div class="user">
		<img :src="gravatar()" class="avatar" alt="" v-tooltip="user.username"/>
		<span v-if="showUsername" class="username">{{ user.username }}</span>
	</div>
</template>

<script>
	export default {
		name: 'user',
		props: {
			user: {
				required: true,
				type: Object,
			},
			showUsername: {
				required: false,
				type: Boolean,
				default: true,
			},
			avatarSize: {
				required: false,
				type: Number,
				default: 50,
			}
		},
		methods: {
			gravatar() {
				return 'https://www.gravatar.com/avatar/' + this.user.avatarUrl + '?s=' + this.avatarSize
			}
		}
	}
</script>

<style lang="scss" scoped>
	.user {
		margin: .5em;

		img {
			-webkit-border-radius: 100%;
			-moz-border-radius: 100%;
			border-radius: 100%;

			vertical-align: middle;
			margin-right: .5em;
		}
	}
</style>
