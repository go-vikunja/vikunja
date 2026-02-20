<template>
	<Modal
		:overflow="true"
		:wide="wide"
		@close="$router.back()"
	>
		<Card
			:title="title"
			:shadow="false"
			:padding="false"
			class="has-text-start"
			:loading="currentLoading"
			:show-close="true"
			@close="$router.back()"
		>
			<div class="p-4">
				<slot />
			</div>

			<template #footer>
				<slot name="footer">
					<XButton
						v-if="tertiary !== ''"
						:shadow="false"
						variant="tertiary"
						@click.prevent.stop="$emit('tertiary', $event)"
					>
						{{ tertiary }}
					</XButton>
					<XButton
						variant="secondary"
						@click.prevent.stop="$router.back()"
					>
						{{ $t('misc.cancel') }}
					</XButton>
					<XButton
						v-if="hasPrimaryAction"
						variant="primary"
						:disabled="isBusy"
						class="mis-2"
						:loading="currentLoading"
						@click.prevent.stop="primary"
					>
						<template
							v-if="showPrimaryIcon"
							#icon
						>
							<slot name="primary-icon">
								<PhPlus />
							</slot>
						</template>
						{{ primaryLabel || $t('misc.create') }}
					</XButton>
				</slot>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import {PhPlus} from '@phosphor-icons/vue'

import {computed, ref, toRef, watch} from 'vue'

const props = withDefaults(defineProps<{
	title: string,
	primaryLabel?: string,
	showPrimaryIcon?: boolean,
	primaryDisabled?: boolean,
	hasPrimaryAction?: boolean,
	tertiary?: string,
	wide?: boolean,
	loading?: boolean,
}>(), {
	primaryLabel: '',
	showPrimaryIcon: true,
	primaryDisabled: false,
	hasPrimaryAction: true,
	tertiary: '',
	wide: false,
})

const emit = defineEmits<{
	'create': [event: MouseEvent],
	'primary': [event: MouseEvent],
	'tertiary': [event: MouseEvent],
	'update:loading': [value: boolean],
}>()

const loadingProp = toRef(props, 'loading')
const currentLoading = ref(false)

watch(
	loadingProp,
	(value) => {
		if (value !== undefined) {
			currentLoading.value = value
		}
	},
	{immediate: true},
)

const isBusy = computed(() => props.primaryDisabled || currentLoading.value)

function setLoading(value: boolean) {
	currentLoading.value = value
	emit('update:loading', value)
}

function primary(event: MouseEvent) {
	if (isBusy.value) {
		return
	}

	emit('create', event)
	emit('primary', event)
	setLoading(true)
}
</script>
