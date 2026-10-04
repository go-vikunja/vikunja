<template>
	<Modal
		variant="hint-modal"
		@close="emit('close')"
	>
		<Card
			class="has-no-shadow"
			:title="$t('risks.statusDialog.title')"
		>
			<p class="mbe-4">
				{{ risk.title }}
			</p>

			<FormField :label="$t('risks.statusDialog.newStatus')">
				<template #default="{id}">
					<FormSelect
						:id="id"
						v-model="target"
						:options="targetOptions"
					/>
				</template>
			</FormField>

			<FormField :label="noteLabel">
				<template #default="{id}">
					<textarea
						:id="id"
						v-model="note"
						class="textarea"
						rows="3"
						maxlength="20000"
					/>
				</template>
			</FormField>
			<p
				v-if="noteKind === 'close'"
				class="help"
			>
				{{ $t('risks.statusDialog.closeNoteHelp') }}
			</p>

			<template #footer>
				<XButton
					variant="tertiary"
					@click="emit('close')"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="primary"
					:disabled="pending"
					:loading="pending"
					@click="confirm"
				>
					{{ confirmLabel }}
				</XButton>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useMutation} from '@tanstack/vue-query'

import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import FormSelect from '@/components/input/FormSelect.vue'

import {changeRiskStatusMutation, isRiskStatusConflict} from '@/client/queries/risks'
import type {Risk} from '@/client/queries/risks'
import {riskNoteKind, statusTargets} from '@/helpers/riskRating'
import type {RiskStatus} from '@/helpers/riskRating'
import {error, success} from '@/message'

const props = defineProps<{
	risk: Risk
	// The status the dialog starts on, for example closed from "Close risk".
	initialStatus?: RiskStatus
}>()

const emit = defineEmits<{
	close: []
	changed: [risk: Risk]
}>()

const {t} = useI18n({useScope: 'global'})

const targets = computed(() => statusTargets(props.risk.status))
const target = ref<RiskStatus>(
	props.initialStatus && props.initialStatus !== props.risk.status ? props.initialStatus : targets.value[0],
)
const note = ref('')

const targetOptions = computed(() => targets.value.map(status => ({value: status, label: t(`risks.statuses.${status}`)})))
const noteKind = computed(() => riskNoteKind(props.risk.status, target.value))

const noteLabel = computed(() => {
	switch (noteKind.value) {
		case 'close': return t('risks.statusDialog.closeNoteLabel')
		case 'reopen': return t('risks.statusDialog.reopenNoteLabel')
		default: return t('risks.statusDialog.noteLabel')
	}
})

const confirmLabel = computed(() => {
	switch (noteKind.value) {
		case 'close': return t('risks.statusDialog.close')
		case 'reopen': return t('risks.statusDialog.reopen')
		default: return t('risks.statusDialog.confirm')
	}
})

const mutation = useMutation(changeRiskStatusMutation())
const pending = computed(() => mutation.isPending.value)

async function confirm() {
	if (pending.value) return
	try {
		const updated = await mutation.mutateAsync({id: props.risk.id, status: target.value, note: note.value.trim()})
		switch (noteKind.value) {
			case 'close': success({message: t('risks.statusDialog.closed')}); break
			case 'reopen': success({message: t('risks.statusDialog.reopened')}); break
			default: success({message: t('risks.statusDialog.changed', {status: t(`risks.statuses.${target.value}`)})})
		}
		emit('changed', updated)
		emit('close')
	} catch (e) {
		if (isRiskStatusConflict(e)) {
			// The mutation already invalidated the risk queries, so the caller shows the current state.
			error({message: t('risks.statusDialog.conflict')})
			emit('close')
			return
		}
		error(e)
	}
}
</script>
