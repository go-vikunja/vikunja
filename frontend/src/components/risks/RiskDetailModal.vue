<template>
	<Modal
		variant="hint-modal"
		@close="emit('close')"
	>
		<Card
			class="has-no-shadow risk-detail"
			:title="`#${risk.id} ${risk.title}`"
		>
			<Message
				v-if="permissionKnown && !writable"
				class="mbe-4"
			>
				{{ $t('risks.detail.readOnly') }}
			</Message>

			<div class="risk-detail__summary">
				<RiskRatingBadge
					:rating="risk.rating"
					:score="risk.score"
				/>
				<span class="tag">{{ $t(`risks.statuses.${risk.status}`) }}</span>
				<span
					v-if="overdue"
					class="tag is-danger"
				>{{ $t('risks.overdue') }}</span>
			</div>

			<dl class="risk-detail__fields">
				<dt>{{ $t('risks.category') }}</dt>
				<dd>{{ risk.category || '-' }}</dd>

				<dt>{{ $t('risks.probability') }}</dt>
				<dd>{{ $t(`risks.levels.probability.${risk.probability}`) }}</dd>

				<dt>{{ $t('risks.impact') }}</dt>
				<dd>{{ $t(`risks.levels.impact.${risk.impact}`) }}</dd>

				<dt>{{ $t('risks.owner') }}</dt>
				<dd>{{ risk.owner ? userName(risk.owner) : $t('risks.noOwner') }}</dd>

				<dt>{{ $t('risks.identifiedDate') }}</dt>
				<dd>{{ formatDay(risk.identified_date) || '-' }}</dd>

				<dt>{{ $t('risks.dueDate') }}</dt>
				<dd>{{ formatDay(risk.due_date) || '-' }}</dd>

				<template v-if="risk.closed_at">
					<dt>{{ $t('risks.closedAt') }}</dt>
					<dd>
						{{ formatDay(risk.closed_at) }}
						<span v-if="risk.closed_by">{{ $t('risks.detail.by', {name: userName(risk.closed_by)}) }}</span>
					</dd>
				</template>

				<template v-if="risk.resolution">
					<dt>{{ $t('risks.resolution') }}</dt>
					<dd class="risk-detail__text">
						{{ risk.resolution }}
					</dd>
				</template>

				<dt>{{ $t('risks.description') }}</dt>
				<dd class="risk-detail__text">
					{{ risk.description || '-' }}
				</dd>

				<dt>{{ $t('risks.mitigation') }}</dt>
				<dd class="risk-detail__text">
					{{ risk.mitigation || '-' }}
				</dd>

				<dt>{{ $t('risks.contingency') }}</dt>
				<dd class="risk-detail__text">
					{{ risk.contingency || '-' }}
				</dd>

				<dt>{{ $t('risks.createdBy') }}</dt>
				<dd>
					{{ risk.created_by ? userName(risk.created_by) : $t('risks.detail.unknownUser') }},
					{{ formatDateTime(risk.created) }}
				</dd>
			</dl>

			<h3 class="risk-detail__heading">
				{{ $t('risks.detail.history') }}
			</h3>
			<p v-if="history.isPending.value">
				{{ $t('misc.loading') }}
			</p>
			<p v-else-if="(history.data.value ?? []).length === 0">
				{{ $t('risks.detail.historyEmpty') }}
			</p>
			<ol
				v-else
				class="risk-detail__history"
			>
				<li
					v-for="entry in history.data.value"
					:key="entry.id"
				>
					<strong>{{ historyTitle(entry) }}</strong>
					<span class="risk-detail__history-meta">
						{{ $t('risks.detail.by', {name: entry.changed_by ? userName(entry.changed_by) : $t('risks.detail.unknownUser')}) }},
						{{ formatDateTime(entry.created) }}
					</span>
					<p
						v-if="entry.note"
						class="risk-detail__text"
					>
						{{ entry.note }}
					</p>
				</li>
			</ol>

			<template #footer>
				<XButton
					variant="tertiary"
					@click="emit('close')"
				>
					{{ $t('misc.close') }}
				</XButton>
				<template v-if="writable">
					<XButton
						variant="secondary"
						:danger="true"
						@click="confirmDelete = true"
					>
						{{ $t('risks.detail.delete') }}
					</XButton>
					<XButton
						variant="secondary"
						@click="editing = true"
					>
						{{ $t('risks.detail.edit') }}
					</XButton>
					<XButton
						v-if="risk.status === 'closed'"
						variant="secondary"
						:disabled="reopening"
						:loading="reopening"
						@click="reopen"
					>
						{{ $t('risks.statusDialog.reopen') }}
					</XButton>
					<XButton
						v-else
						variant="secondary"
						@click="openStatus('closed')"
					>
						{{ $t('risks.statusDialog.close') }}
					</XButton>
					<XButton
						variant="primary"
						@click="openStatus()"
					>
						{{ $t('risks.statusDialog.change') }}
					</XButton>
				</template>
			</template>
		</Card>
	</Modal>

	<RiskFormModal
		v-if="editing"
		:risk="risk"
		@close="editing = false"
	/>

	<RiskStatusDialog
		v-if="statusOpen"
		:risk="risk"
		:initial-status="statusInitial"
		@close="statusOpen = false"
	/>

	<Modal
		v-if="confirmDelete"
		variant="hint-modal"
		@close="confirmDelete = false"
	>
		<Card
			class="has-no-shadow"
			:title="$t('risks.detail.deleteTitle')"
		>
			<p>{{ $t('risks.detail.deleteBody') }}</p>
			<template #footer>
				<XButton
					variant="tertiary"
					@click="confirmDelete = false"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="primary"
					:danger="true"
					:disabled="deleting"
					:loading="deleting"
					@click="remove"
				>
					{{ $t('risks.detail.delete') }}
				</XButton>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useMutation, useQuery} from '@tanstack/vue-query'
import dayjs from 'dayjs'

import Card from '@/components/misc/Card.vue'
import Message from '@/components/misc/Message.vue'
import Modal from '@/components/misc/Modal.vue'
import XButton from '@/components/input/Button.vue'
import RiskFormModal from '@/components/risks/RiskFormModal.vue'
import RiskRatingBadge from '@/components/risks/RiskRatingBadge.vue'
import RiskStatusDialog from '@/components/risks/RiskStatusDialog.vue'

import {
	canWriteRisk,
	changeRiskStatusMutation,
	deleteRiskMutation,
	isRiskStatusConflict,
	riskHistoryQuery,
	riskQuery,
} from '@/client/queries/risks'
import type {Risk, RiskHistoryEntry, RiskUser} from '@/client/queries/risks'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {isRiskOverdue, REOPEN_STATUS} from '@/helpers/riskRating'
import type {RiskStatus} from '@/helpers/riskRating'
import {error, success} from '@/message'

const props = defineProps<{
	// The risk as the list had it. The modal loads the current one, with what the user may do with it.
	initial: Risk
}>()

const emit = defineEmits<{
	close: []
	deleted: []
}>()

const {t} = useI18n({useScope: 'global'})

// The data of the list is shown at once and replaced when the single risk arrives.
// Once the risk is being deleted its queries stop, so the refresh after the delete does not ask for a risk that is gone.
const gone = ref(false)
const query = useQuery(computed(() => ({
	...riskQuery(props.initial.id),
	initialData: props.initial,
	initialDataUpdatedAt: 0,
	enabled: !gone.value,
})))
const risk = computed<Risk>(() => query.data.value ?? props.initial)
const history = useQuery(computed(() => ({...riskHistoryQuery(props.initial.id), enabled: !gone.value})))

// A list item carries no permission. Until the single risk arrived nothing can be changed.
const permissionKnown = computed(() => risk.value.max_permission !== undefined)
const writable = computed(() => canWriteRisk(risk.value))
const overdue = computed(() => isRiskOverdue(risk.value.status, risk.value.due_date))

function userName(user: RiskUser): string {
	return user.name || user.username
}

function formatDay(value: string | null): string {
	const date = parseDateOrNull(value)
	return date ? dayjs(date).format('L') : ''
}

function formatDateTime(value: string): string {
	const date = parseDateOrNull(value)
	return date ? dayjs(date).format('L LT') : ''
}

function historyTitle(entry: RiskHistoryEntry): string {
	const to = t(`risks.statuses.${entry.to_status}`)
	return entry.from_status === ''
		? t('risks.detail.historyCreated', {to})
		: t('risks.detail.historyChanged', {from: t(`risks.statuses.${entry.from_status}`), to})
}

// ---- edit and status -----------------------------------------------------------------------
const editing = ref(false)
const statusOpen = ref(false)
const statusInitial = ref<RiskStatus | undefined>(undefined)

function openStatus(initial?: RiskStatus) {
	statusInitial.value = initial
	statusOpen.value = true
}

// Reopening is one click: back to open, no note. The status dialog can reopen as something else.
const reopenMutation = useMutation(changeRiskStatusMutation())
const reopening = computed(() => reopenMutation.isPending.value)

async function reopen() {
	if (reopening.value) return
	try {
		await reopenMutation.mutateAsync({id: risk.value.id, status: REOPEN_STATUS})
		success({message: t('risks.statusDialog.reopened')})
	} catch (e) {
		if (isRiskStatusConflict(e)) {
			error({message: t('risks.statusDialog.conflict')})
			return
		}
		error(e)
	}
}

// ---- delete --------------------------------------------------------------------------------
const confirmDelete = ref(false)
const deleteMutation = useMutation(deleteRiskMutation())
const deleting = computed(() => deleteMutation.isPending.value)

async function remove() {
	if (deleting.value) return
	gone.value = true
	try {
		await deleteMutation.mutateAsync(risk.value.id)
		success({message: t('risks.detail.deleted')})
		confirmDelete.value = false
		emit('deleted')
		emit('close')
	} catch (e) {
		gone.value = false
		error(e)
	}
}
</script>

<style lang="scss" scoped>
.risk-detail__summary {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .5rem;
	margin-block-end: 1rem;
}

.risk-detail__fields {
	display: grid;
	grid-template-columns: max-content 1fr;
	gap: .5rem 1rem;
	margin: 0 0 1.5rem;

	dt {
		font-weight: 600;
	}

	dd {
		margin: 0;
	}

	@media screen and (max-width: $mobile) {
		grid-template-columns: 1fr;
		gap: 0;

		dd {
			margin-block-end: .5rem;
		}
	}
}

.risk-detail__text {
	white-space: pre-wrap;
	overflow-wrap: anywhere;
}

.risk-detail__heading {
	font-weight: 600;
	margin-block-end: .5rem;
}

.risk-detail__history {
	list-style: none;
	margin: 0;
	padding: 0;

	li {
		padding-block: .5rem;
		border-block-start: 1px solid var(--grey-200);
	}
}

.risk-detail__history-meta {
	display: block;
	font-size: .85rem;
	color: var(--grey-500);
}
</style>
