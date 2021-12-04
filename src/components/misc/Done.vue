<template>
    <div
		v-if="isDone"
		class="is-done"
		:class="{ 'is-done--small': variant === 'small' }"
	>
		{{ $t('task.attributes.done') }}
	</div>
</template>

<script lang="ts" setup>
import {PropType} from 'vue'
type Variants = 'default' | 'small'

defineProps({
	isDone: {
		type: Boolean,
		default: false,
	},
	variant: {
		type: String as PropType<Variants>,
		default: 'default',
	},
})
</script>


<style lang="scss" scoped>
.is-done {
  background: var(--success);
  color: var(--white);
  padding: .5rem;
  font-weight: bold;
  line-height: 1;
  border-radius: 4px;
  text-align: center;
}

.is-done--small {
	padding: .2rem .3rem;
}
</style>