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
withDefaults(defineProps<{
	value: number
	isSmall?: boolean
	isPrimary?: boolean
}>(), {
	isSmall: false,
	isPrimary: false,
})
</script>

<style lang="scss" scoped>
@use 'sass:list';

.progress-bar {
	--progress-height: #{$size-normal};
	--progress-bar-background-color: var(--border-light, #{$border-light});
	--progress-value-background-color: var(--grey-500, #{$text});
	--progress-border-radius: #{$radius};
	--progress-indeterminate-duration: 1.5s;

	appearance: none;
	border: none;
	border-radius: var(--progress-border-radius);
	height: var(--progress-height);
	overflow: hidden;
	padding: 0;
	min-width: 6vw;

	width: 50px;
	margin: 0 .5rem 0 0;
	flex: 3 1 auto;

	&::-moz-progress-bar,
	&::-webkit-progress-value {
		background: var(--progress-value-background-color);
	}

	@media screen and (max-width: $tablet) {
		margin: 0.5rem 0 0 0;
		order: 1;
		width: 100%;
	}

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
		$color: list.nth($pair, 1);
		&.is-#{$name} {
			--progress-value-background-color: var(--#{$name}, #{$color});

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

	&.is-small {
		--progress-height: #{$size-small};
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
</style>
