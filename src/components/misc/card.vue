<template>
	<div class="card" :class="{'has-no-shadow': !shadow}">
		<header class="card-header" v-if="title !== ''">
			<p class="card-header-title">
				{{ title }}
			</p>
			<a @click="$emit('close')" class="card-header-icon" v-if="hasClose">
				<span class="icon">
					<icon :icon="closeIcon"/>
				</span>
			</a>
		</header>
		<div class="card-content" :class="{'p-0': !padding}">
			<div class="content">
				<slot></slot>
			</div>
		</div>
	</div>
</template>

<script>
export default {
	name: 'card',
	props: {
		title: {
			type: String,
			default: '',
		},
		padding: {
			type: Boolean,
			default: true,
		},
		hasClose: {
			type: Boolean,
			default: false,
		},
		closeIcon: {
			type: String,
			default: 'angle-right',
		},
		shadow: {
			type: Boolean,
			default: true,
		},
	},
}
</script>
