<template>
	<transition
		name="expandable-slide"
		@beforeEnter="beforeEnter"
		@enter="enter"
		@afterEnter="afterEnter"
		@enterCancelled="enterCancelled"
		@beforeLeave="beforeLeave"
		@leave="leave"
		@afterLeave="afterLeave"
		@leaveCancelled="leaveCancelled"
	>
		<div
			v-if="initialHeight"
			class="expandable-initial-height"
			:style="{ maxHeight: `${initialHeight}px` }"
			:class="{ 'expandable-initial-height--expanded': open }"
		>
			<slot />
		</div>
		<div
			v-else-if="open"
			class="expandable"
		>
			<slot />
		</div>
	</transition>
</template>

<script setup lang="ts">
// the logic of this component is loosely based on this article
// https://gomakethings.com/how-to-add-transition-animations-to-vanilla-javascript-show-and-hide-methods/#putting-it-all-together

import {computed, ref} from 'vue'
import {getInheritedBackgroundColor} from '@/helpers/getInheritedBackgroundColor'

const props = withDefaults(defineProps<{
  /** Whether the Expandable is open or not */
  open?: boolean,
  /** If there is too much content, content will be cut of here. */
  initialHeight?: number
  /** The hidden content is indicated by a gradient. This is the color that the gradient fades to.
  * Makes only sense if `initialHeight` is set. */
	backgroundColor?: string
}>(), {
	open: false,
	initialHeight: undefined,
	backgroundColor: undefined,
})

const wrapper = ref<HTMLElement | null>(null)

const computedBackgroundColor = computed(() => {
	if (wrapper.value === null) {
		return props.backgroundColor || '#fff'
	}
	return props.backgroundColor || getInheritedBackgroundColor(wrapper.value)
})

/**
 * Get the natural height of the element
 */
function getHeight(el: HTMLElement) {
	const { display } = el.style // save display property
	el.style.display = 'block' // Make it visible
	const height = `${el.scrollHeight}px` // Get its height
	el.style.display = display // revert to original display property
	return height
}

/**
 * force layout of element changes
 * https://gist.github.com/paulirish/5d52fb081b3570c81e3a
 */
function forceLayout(el: HTMLElement) {
	// eslint-disable-next-line @typescript-eslint/no-unused-expressions
	el.offsetTop
}

/* ######################################################################
# The following functions are called by the js hooks of the transitions.
# They follow the original hook order of the vue transition component
# see: https://vuejs.org/guide/built-ins/transition.html#javascript-hooks
###################################################################### */

function beforeEnter(el: HTMLElement) {
	el.style.height = '0'
	el.style.willChange = 'height'
	el.style.backfaceVisibility = 'hidden'
	forceLayout(el)
}

// the done callback is optional when
// used in combination with CSS
function enter(el: HTMLElement) {
	const height = getHeight(el) // Get the natural height
	el.style.height = height // Update the height
}

function afterEnter(el: HTMLElement) {
	removeHeight(el)
}

function enterCancelled(el: HTMLElement) {
	removeHeight(el)
}

function beforeLeave(el: HTMLElement) {
	// Give the element a height to change from
	el.style.height = `${el.scrollHeight}px`
	forceLayout(el)
}

function leave(el: HTMLElement) {
	// Set the height back to 0
	el.style.height = '0'
	el.style.willChange = ''
	el.style.backfaceVisibility = ''
}

function afterLeave(el: HTMLElement) {
	removeHeight(el)
}

function leaveCancelled(el: HTMLElement) {
	removeHeight(el)
}

function removeHeight(el: HTMLElement) {
	el.style.height = ''
}
</script>

<style lang="scss" scoped>
$transition-time: 300ms;

// https://easings.net/#easeInQuint
$ease-in-quint: cubic-bezier(0.64, 0, 0.78, 0);
// https://easings.net/#easeInOutQuint
$ease-in-out-quint: cubic-bezier(0.83, 0, 0.17, 1);

.expandable-slide-enter-active,
.expandable-slide-leave-active {
  transition:
    opacity $transition-time $ease-in-quint,
    height $transition-time $ease-in-out-quint;
  overflow: hidden;
}

.expandable-slide-enter,
.expandable-slide-leave-to {
  opacity: 0;
}

.expandable-initial-height {
  padding: 5px;
  margin: -5px;
  overflow: hidden;
  position: relative;

  &::after {
    content: "";
    display: block;
    background-image: linear-gradient(
      to bottom,
      transparent,
      ease-in-out
      v-bind(computedBackgroundColor)
    );
    position: absolute;
    block-size: 40px;
    inline-size: 100%;
    inset-block-end: 0;
  }
}

.expandable-initial-height--expanded {
  block-size: 100% !important;

  &::after {
    display: none;
  }
}
</style>
