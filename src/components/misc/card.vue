<template>
	<div class="card" :class="{'has-no-shadow': !shadow}">
		<header class="card-header" v-if="title !== ''">
			<p class="card-header-title">
				{{ title }}
			</p>
			<a
				v-if="hasClose"
				class="card-header-icon"
				:aria-label="$t('misc.close')"
				@click="$emit('close')"
				v-tooltip="$t('misc.close')"
			>	
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

<script setup lang="ts">
defineProps({
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
		default: 'times',
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
})

defineEmits(['close'])
</script>

<style lang="scss" scoped>
.card {
  background-color: var(--white);
  border-radius: $radius;
  margin-bottom: 1rem;
  border: 1px solid var(--card-border-color);
  box-shadow: var(--shadow-sm);
}

.card-header {
  box-shadow: none;
  border-bottom: 1px solid var(--card-border-color);
  border-radius: $radius $radius 0 0;
}

// FIXME: should maybe be merged somehow with modal
:deep(.modal-card-foot) {
  background-color: var(--grey-50);
  border-top: 0;
}
</style>