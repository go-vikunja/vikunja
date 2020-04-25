<template>
	<div class="user" :class="{'is-inline': isInline}">
		<img :src="user.getAvatarUrl(avatarSize)" class="avatar" alt="" v-tooltip="user.username" :width="avatarSize" :height="avatarSize"/>
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
			},
			isInline: {
				required: false,
				type: Boolean,
				default: false,
			},
		},
	}
</script>

<style lang="scss" scoped>
	.user {
		margin: .5em;

		&.is-inline {
			display: inline;
		}

		img {
			-webkit-border-radius: 100%;
			-moz-border-radius: 100%;
			border-radius: 100%;

			vertical-align: middle;
			margin-right: .5em;
		}
	}
</style>
