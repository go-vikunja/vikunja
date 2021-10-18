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
		<div class="card-content loader-container" :class="{'p-0': !padding, 'is-loading': loading}">
			<div :class="{'content': hasContent}">
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
		hasContent: {
			type: Boolean,
			default: true,
		},
		loading: {
			type: Boolean,
			default: false,
		},
	},
	emits: ['close'],
}
</script>

<style lang="scss" scoped>
.card {
  background-color: $white;
  border-radius: $radius;
  margin-bottom: 1rem;
  border: 1px solid $grey-200;
  box-shadow: $shadow-sm;
}

.card-header {
  box-shadow: none;
  border-bottom: 1px solid $grey-200;
  border-radius: $radius $radius 0 0;
}

// FIXME: should maybe be merged somehow with modal
::v-deep.modal-card-foot {
  background-color: $grey-50;
  border-top: 0;
}
</style>