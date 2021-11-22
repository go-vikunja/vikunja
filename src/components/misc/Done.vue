<template>
    <div
			v-if="isDone"
			class="is-done"
			:class="{ 'is-done--small': variant === variants.SMALL }"
		>
			{{ $t('task.attributes.done') }}
		</div>
</template>

<script>
const VARIANTS = {
	DEFAULT: 'default',
	SMALL: 'small',
}

export default {
	name: 'Done',

	data() {
		return {
			variants: VARIANTS,
		}
	},

	props: {
		isDone: {
			type: Boolean,
			default: false,
		},
		variant: {
			type: String,
			default: VARIANTS.DEFAULT,
			validator: (variant) => Object.values(VARIANTS).includes(variant),
		},
	},
}
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