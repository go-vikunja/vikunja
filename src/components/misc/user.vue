<template>
	<div :class="{'is-inline': isInline}" class="user">
		<img
			:height="avatarSize"
			:src="user.getAvatarUrl(avatarSize)"
			:width="avatarSize"
			alt=""
			class="avatar"
			v-tooltip="user.getDisplayName()"/>
		<span class="username" v-if="showUsername">{{ user.getDisplayName() }}</span>
	</div>
</template>

<script lang="ts" setup>
defineProps({
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
})
</script>

<style lang="scss" scoped>
.user {
	margin: .5rem;

	&.is-inline {
		display: inline;
	}

	img {
		border-radius: 100%;

		vertical-align: middle;
		margin-right: .5rem;
	}
}
</style>
