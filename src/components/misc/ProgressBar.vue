<template>
	<progress
		class="progress-bar"
		:class="{
			'is-small': isSmall,
			'is-primary': isPrimary,
		}"
		:value="value"
		max="100"
	>
		{{ value }}%
	</progress>
</template>

<script setup lang="ts">
import {defineProps} from 'vue'

defineProps({
	value: {
		type: Number,
		required: true,
	},
	isSmall: {
		type: Boolean,
		default: false,
	},
	isPrimary: {
		type: Boolean,
		required: false,
	},
})
</script>

<style lang="scss" scoped>
.progress-bar {
	--progress-height: var(--size-normal, #{$size-normal});
	--progress-bar-background-color: var(--border-light, #{$border-light});
	--progress-value-background-color: var(--text, #{$text});
	--progress-border-radius: var(--radius-rounded, #{$radius-rounded});
	--progress-indeterminate-duration: 1.5s;

	--size-small: #{$size-small};
	--size-medium: #{$size-medium};
	--size-large: #{$size-large};

	appearance: none;
	border: none;
	border-radius: var(--progress-border-radius);
	display: block;
	height: var(--progress-height);
	overflow: hidden;
	padding: 0;
	width: 100%;

	&::-webkit-progress-bar {
		background-color: var(--progress-bar-background-color);
	}
	&::-webkit-progress-value {
		background-color: var(--progress-value-background-color);
	}
	&::-moz-progress-bar {
		background-color: var(--progress-value-background-color);
	}
	&::-ms-fill {
		background-color: var(--progress-value-background-color);
		border: none;
	}

	// Colors
	@each $name, $pair in $colors {
		$color: nth($pair, 1);
		&.is-#{$name} {
			&::-webkit-progress-value {
				--progress-value-background-color: var(--#{$name}, #{$color});
			}

			&::-moz-progress-bar {
				--progress-value-background-color: var(--#{$name}, #{$color});
			}

			&::-ms-fill {
				--progress-value-background-color: var(--#{$name}, #{$color});
			}

			&:indeterminate {
				background-image: linear-gradient(
					to right,
					var(--#{$name}, #{$color}) 30%,
					var(--progress-bar-background-color) 30%
				);
			}
		}
	}

	&:indeterminate {
		animation-duration: var(--progress-indeterminate-duration);
		animation-iteration-count: infinite;
		animation-name: moveIndeterminate;
		animation-timing-function: linear;
		background-color: var(--progress-bar-background-color);
		background-image: linear-gradient(
			to right,
			var(--text, #{$text}) 30%,
			var(--progress-bar-background-color) 30%
		);
		background-position: top left;
		background-repeat: no-repeat;
		background-size: 150% 150%;

		&::-webkit-progress-bar {
			background-color: transparent;
		}

		&::-moz-progress-bar {
			background-color: transparent;
		}

		&::-ms-fill {
			animation-name: none;
		}
	}

	// Sizes
	&.is-small {
		--progress-height: var(--size-small, #{$size-small});
	}
	&.is-medium {
		--progress-height: var(--size-medium, #{$size-medium});
	}
	&.is-large {
		--progress-height: var(--size-large, #{$size-large});
	}
}

@keyframes moveIndeterminate {
	from {
		background-position: 200% 0;
	}
	to {
		background-position: -200% 0;
	}
}

.progress-bar {
	--progress-height: var(--size-normal, 1rem);

	border-radius: $radius-large;
	min-width: 6vw;

	@media screen and (max-width: $tablet) {
		width: 100%;
	}

	&::-moz-progress-bar,
	&::-webkit-progress-value {
		background: var(--grey-500);
	}
}

.progress-bar.is-small {
	--progress-height: var(--size-small, 0.75rem);
}
</style>
