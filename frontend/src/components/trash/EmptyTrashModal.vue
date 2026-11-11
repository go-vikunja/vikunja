<template>
	<Modal
		:enabled="enabled"
		@close="emit('close')"
		@submit="emit('submit')"
	>
		<template #header>
			<span>{{ $t('trash.empty') }}</span>
		</template>
		<template #text>
			<Loading
				v-if="!count"
				variant="small"
			/>
			<template v-else>
				<p
					v-cy="'emptyTrashText'"
					class="tw:text-balance"
				>
					{{ $t(texts.text, purgedCount) }}
				</p>
				<template v-if="(count.by_others ?? 0) > 0">
					<Message
						v-cy="'emptyTrashOthersWarning'"
						variant="warning"
						class="mbe-3"
					>
						{{ $t(texts.byOthers, count.by_others ?? 0) }}
					</Message>
					<FancyCheckbox
						v-model="includeOthers"
						v-cy="'emptyTrashIncludeOthers'"
					>
						{{ $t(texts.includeOthers) }}
					</FancyCheckbox>
				</template>
			</template>
		</template>
	</Modal>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import Loading from '@/components/misc/Loading.vue'
import Message from '@/components/misc/Message.vue'
import Modal from '@/components/misc/Modal.vue'
import type {TrashCount} from '@/client/generated'

const props = defineProps<{
	enabled: boolean,
	count: TrashCount | undefined,
	// Pluralised by count: text, byOthers
	texts: {text: string, byOthers: string, includeOthers: string},
}>()

const emit = defineEmits<{
	close: [],
	submit: [],
}>()

const includeOthers = defineModel<boolean>('includeOthers', {required: true})

const purgedCount = computed(() => {
	const total = props.count?.total ?? 0
	return includeOthers.value ? total : total - (props.count?.by_others ?? 0)
})
</script>
