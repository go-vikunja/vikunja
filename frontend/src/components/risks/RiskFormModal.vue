<template>
	<Modal
		variant="hint-modal"
		@close="emit('close')"
	>
		<Card
			class="has-no-shadow"
			:title="risk ? $t('risks.form.editTitle') : $t('risks.form.createTitle')"
		>
			<FormField
				v-if="!risk && projectId === null"
				:label="$t('risks.project')"
			>
				<template #default="{id}">
					<FormSelect
						:id="id"
						v-model.number="chosenProjectId"
						:options="projectOptions"
					/>
				</template>
			</FormField>

			<FormField
				:label="$t('risks.titleField')"
				:error="titleError"
			>
				<template #default="{id}">
					<FormInput
						:id="id"
						v-model="form.title"
						type="text"
						maxlength="250"
						required
					/>
				</template>
			</FormField>

			<FormField :label="$t('risks.category')">
				<template #default="{id}">
					<FormInput
						:id="id"
						v-model="form.category"
						type="text"
						maxlength="100"
					/>
				</template>
			</FormField>

			<div class="risk-form__levels">
				<FormField :label="$t('risks.probability')">
					<template #default="{id}">
						<FormSelect
							:id="id"
							v-model.number="form.probability"
							:options="levelOptions('probability')"
						/>
					</template>
				</FormField>
				<FormField :label="$t('risks.impact')">
					<template #default="{id}">
						<FormSelect
							:id="id"
							v-model.number="form.impact"
							:options="levelOptions('impact')"
						/>
					</template>
				</FormField>
			</div>
			<p
				class="risk-form__preview"
				aria-live="polite"
			>
				<RiskRatingBadge
					v-if="previewRating"
					:rating="previewRating"
				/>
				{{ $t('risks.scorePreview', {score: previewScore, rating: previewRating ? $t(`risks.ratings.${previewRating}`) : ''}) }}
			</p>

			<FormField :label="$t('risks.description')">
				<template #default="{id}">
					<textarea
						:id="id"
						v-model="form.description"
						class="textarea"
						rows="3"
					/>
				</template>
			</FormField>

			<FormField :label="$t('risks.owner')">
				<template #default="{id}">
					<Multiselect
						:id="id"
						v-model="owner"
						:loading="ownerLoading"
						:placeholder="$t('risks.noOwner')"
						:search-results="ownerResults"
						:disabled="!effectiveProjectId"
						label="username"
						@search="searchOwners"
					>
						<template #searchResult="{option}">
							<span>{{ option.name || option.username }} <small>({{ option.username }})</small></span>
						</template>
					</Multiselect>
				</template>
			</FormField>
			<p class="help">
				{{ $t('risks.form.ownerHelp') }}
			</p>

			<FormField :label="$t('risks.mitigation')">
				<template #default="{id}">
					<textarea
						:id="id"
						v-model="form.mitigation"
						class="textarea"
						rows="3"
					/>
				</template>
			</FormField>

			<FormField :label="$t('risks.contingency')">
				<template #default="{id}">
					<textarea
						:id="id"
						v-model="form.contingency"
						class="textarea"
						rows="3"
					/>
				</template>
			</FormField>

			<div class="risk-form__levels">
				<FormField :label="$t('risks.identifiedDate')">
					<template #default="{id}">
						<FormInput
							:id="id"
							v-model="form.identified"
							type="date"
						/>
					</template>
				</FormField>
				<FormField
					:label="$t('risks.dueDate')"
					:error="dateError"
				>
					<template #default="{id}">
						<FormInput
							:id="id"
							v-model="form.due"
							type="date"
						/>
					</template>
				</FormField>
			</div>

			<template #footer>
				<XButton
					variant="tertiary"
					@click="emit('close')"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="primary"
					:disabled="!canSave || saving"
					:loading="saving"
					@click="save"
				>
					{{ $t('risks.form.save') }}
				</XButton>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import {computed, reactive, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useMutation} from '@tanstack/vue-query'

import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import XButton from '@/components/input/Button.vue'
import FormField from '@/components/input/FormField.vue'
import FormInput from '@/components/input/FormInput.vue'
import FormSelect from '@/components/input/FormSelect.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import RiskRatingBadge from '@/components/risks/RiskRatingBadge.vue'

import {createRiskMutation, updateRiskMutation} from '@/client/queries/risks'
import type {Risk, RiskInput} from '@/client/queries/risks'
import {searchProjectUsers} from '@/client/queries/userSearch'
import {dayInputToIso, dueBeforeIdentified, isoToDayInput} from '@/helpers/riskDates'
import {MAX_RISK_LEVEL, MIN_RISK_LEVEL, riskRating, riskScore} from '@/helpers/riskRating'
import {error, success} from '@/message'

type OwnerOption = {id: number, name: string, username: string}

const props = withDefaults(defineProps<{
	// The risk to edit. Without one the form creates a risk.
	risk?: Risk | null
	// The project of a new risk. null lets the user choose one of `projects`.
	projectId?: number | null
	projects?: {id: number, title: string}[]
}>(), {
	risk: null,
	projectId: null,
	projects: () => [],
})

const emit = defineEmits<{
	close: []
	saved: [risk: Risk]
}>()

const {t} = useI18n({useScope: 'global'})

const form = reactive({
	title: props.risk?.title ?? '',
	category: props.risk?.category ?? '',
	description: props.risk?.description ?? '',
	probability: props.risk?.probability ?? 3,
	impact: props.risk?.impact ?? 3,
	mitigation: props.risk?.mitigation ?? '',
	contingency: props.risk?.contingency ?? '',
	identified: props.risk ? isoToDayInput(props.risk.identified_date) : isoToDayInput(new Date().toISOString()),
	due: isoToDayInput(props.risk?.due_date),
})

const chosenProjectId = ref<number>(props.projects[0]?.id ?? 0)
const effectiveProjectId = computed(() => props.risk?.project_id ?? props.projectId ?? chosenProjectId.value)
const projectOptions = computed(() => props.projects.map(p => ({value: p.id, label: p.title})))

function levelOptions(kind: 'probability' | 'impact') {
	const options = []
	for (let level = MIN_RISK_LEVEL; level <= MAX_RISK_LEVEL; level++) {
		options.push({value: level, label: t(`risks.levels.${kind}.${level}`)})
	}
	return options
}

const previewScore = computed(() => riskScore(form.probability, form.impact))
const previewRating = computed(() => riskRating(previewScore.value))

// ---- owner ---------------------------------------------------------------------------------
const owner = ref<OwnerOption | null>(props.risk?.owner
	? {id: props.risk.owner.id, name: props.risk.owner.name, username: props.risk.owner.username}
	: null)
const ownerResults = ref<OwnerOption[]>([])
const ownerLoading = ref(false)

async function searchOwners(query: string) {
	if (!query || !effectiveProjectId.value) {
		ownerResults.value = []
		return
	}
	ownerLoading.value = true
	try {
		const found = await searchProjectUsers(effectiveProjectId.value, query)
		ownerResults.value = found.map(u => ({id: Number(u.id), name: u.name ?? '', username: u.username ?? ''}))
	} catch (e) {
		error(e)
	} finally {
		ownerLoading.value = false
	}
}

// ---- validation ----------------------------------------------------------------------------
const titleError = computed(() => form.title !== '' && form.title.trim() === '' ? t('risks.form.titleRequired') : '')
const dateError = computed(() => dueBeforeIdentified(form.identified, form.due) ? t('risks.form.dueBeforeIdentified') : '')
const canSave = computed(() =>
	form.title.trim() !== '' &&
	!dateError.value &&
	effectiveProjectId.value > 0,
)

// ---- save ----------------------------------------------------------------------------------
const createMutation = useMutation(createRiskMutation())
const updateMutation = useMutation(updateRiskMutation())
const saving = computed(() => createMutation.isPending.value || updateMutation.isPending.value)

function toInput(): RiskInput {
	return {
		title: form.title.trim(),
		description: form.description,
		category: form.category.trim(),
		probability: form.probability,
		impact: form.impact,
		owner_id: owner.value?.id ?? 0,
		mitigation: form.mitigation,
		contingency: form.contingency,
		identified_date: dayInputToIso(form.identified),
		due_date: dayInputToIso(form.due, true),
	}
}

async function save() {
	if (!canSave.value || saving.value) return
	try {
		let saved: Risk
		if (props.risk) {
			saved = await updateMutation.mutateAsync({id: props.risk.id, ...toInput()})
			success({message: t('risks.form.saved')})
		} else {
			saved = await createMutation.mutateAsync({projectId: effectiveProjectId.value, ...toInput()})
			success({message: t('risks.form.created')})
		}
		emit('saved', saved)
		emit('close')
	} catch (e) {
		error(e)
	}
}
</script>

<style lang="scss" scoped>
.risk-form__levels {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: 1rem;

	@media screen and (max-width: $mobile) {
		grid-template-columns: 1fr;
	}
}

.risk-form__preview {
	display: flex;
	align-items: center;
	gap: .5rem;
	margin-block-end: 1rem;
	font-size: .9rem;
}
</style>
