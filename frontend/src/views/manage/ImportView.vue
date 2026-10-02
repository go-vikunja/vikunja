<template>
	<div class="manage-import">
		<Card :title="$t('manage.import.statusTitle')">
			<p v-if="status.isPending.value">
				{{ $t('misc.loading') }}
			</p>
			<template v-else-if="data">
				<dl class="manage-import__meta">
					<dt>{{ $t('manage.import.scheduled') }}</dt>
					<dd>
						{{ data.enabled ? $t('manage.import.scheduleOn', {schedule: data.schedule}) : $t('manage.import.scheduleOff') }}
					</dd>
					<template v-if="data.enabled">
						<dt>{{ $t('manage.import.scheduledMode') }}</dt>
						<dd>{{ data.dry_run ? $t('manage.import.modeDryRun') : $t('manage.import.modeLive') }}</dd>
					</template>
					<dt>{{ $t('manage.import.file') }}</dt>
					<dd>
						<template v-if="data.file_exists">
							{{ formatBytes(data.file_size) }},
							<TimeDisplay :date="data.file_modified" />
						</template>
						<span
							v-else
							class="has-text-danger"
						>{{ $t('manage.import.noFile') }}</span>
					</dd>
					<template v-if="data.last_run">
						<dt>{{ $t('manage.import.lastRun') }}</dt>
						<dd>
							<TimeDisplay :date="data.last_run.finished_at" />
							({{ data.last_run.trigger === 'manual'
								? $t('manage.import.triggerManual', {actor: data.last_run.actor ?? ''})
								: $t('manage.import.triggerSchedule') }})
						</dd>
						<template v-if="data.last_run.error">
							<dt>{{ $t('manage.import.lastError') }}</dt>
							<dd class="has-text-danger">
								{{ data.last_run.error }}
							</dd>
						</template>
						<template v-if="data.last_run.counts">
							<dt>{{ $t('manage.import.lastResult') }}</dt>
							<dd>
								{{ summary(data.last_run.counts) }}
								<span v-if="data.last_run.dry_run">({{ $t('manage.import.modeDryRun') }})</span>
							</dd>
						</template>
					</template>
				</dl>
				<p class="help">
					{{ $t('manage.import.memoryNote') }}
				</p>
			</template>
		</Card>

		<Card
			:title="$t('manage.import.uploadTitle')"
			class="mbs-4"
		>
			<p class="mbe-4">
				{{ $t('manage.import.uploadHelp') }}
			</p>
			<input
				ref="fileInput"
				type="file"
				accept=".csv,text/csv"
				@change="onFileChosen"
			>
			<XButton
				class="mbs-4"
				:disabled="!chosenFile || uploadMutation.isPending.value"
				:loading="uploadMutation.isPending.value"
				@click="upload"
			>
				{{ $t('manage.import.upload') }}
			</XButton>
			<p
				v-if="uploaded"
				class="mbs-4"
			>
				{{ $t('manage.import.uploaded', {rows: uploaded.rows, invalid: uploaded.skipped_invalid, duplicates: uploaded.skipped_duplicate}) }}
			</p>
		</Card>

		<Card
			:title="$t('manage.import.runTitle')"
			class="mbs-4"
		>
			<p class="mbe-4">
				{{ $t('manage.import.runHelp') }}
			</p>
			<div class="buttons">
				<XButton
					variant="secondary"
					:disabled="!data?.file_exists || previewMutation.isPending.value || data?.running"
					:loading="previewMutation.isPending.value"
					@click="preview"
				>
					{{ $t('manage.import.preview') }}
				</XButton>
				<XButton
					variant="primary"
					:disabled="!previewed || !!previewResult?.blocked || data?.running || runMutation.isPending.value"
					:loading="data?.running || runMutation.isPending.value"
					@click="confirmRun = true"
				>
					{{ data?.running ? $t('manage.import.running') : $t('manage.import.run') }}
				</XButton>
			</div>
			<p
				v-if="!previewed && data?.file_exists"
				class="help"
			>
				{{ $t('manage.import.previewFirst') }}
			</p>

			<template v-if="previewResult">
				<Message
					v-if="previewResult.blocked"
					variant="danger"
					class="mbs-4"
				>
					{{ $t('manage.import.blocked', {reason: previewResult.blocked}) }}
				</Message>
				<p class="mbs-4">
					<strong>{{ summary(previewResult.counts) }}</strong>
				</p>
				<template v-if="previewResult.details">
					<ImportEntryList
						:title="$t('manage.import.willDisable')"
						:entries="previewResult.details.disable"
						:total="previewResult.counts.disabled"
					/>
					<ImportEntryList
						:title="$t('manage.import.willCreate')"
						:entries="previewResult.details.create"
						:total="previewResult.counts.created"
					/>
					<ImportEntryList
						:title="$t('manage.import.willUpdate')"
						:entries="previewResult.details.update"
						:total="previewResult.counts.updated"
					/>
					<p
						v-if="previewResult.details.truncated"
						class="help"
					>
						{{ $t('manage.import.truncated') }}
					</p>
				</template>
			</template>
		</Card>
	</div>

	<Modal
		v-if="confirmRun"
		variant="hint-modal"
		@close="confirmRun = false"
	>
		<Card
			class="has-no-shadow"
			:title="$t('manage.import.confirmTitle')"
		>
			<p>
				{{ $t('manage.import.confirmBody', {disabled: previewResult?.counts.disabled ?? 0, created: previewResult?.counts.created ?? 0}) }}
			</p>
			<template #footer>
				<XButton
					variant="tertiary"
					@click="confirmRun = false"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="primary"
					:danger="(previewResult?.counts.disabled ?? 0) > 0"
					:loading="runMutation.isPending.value"
					@click="run"
				>
					{{ $t('manage.import.run') }}
				</XButton>
			</template>
		</Card>
	</Modal>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useMutation, useQuery} from '@tanstack/vue-query'

import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import Message from '@/components/misc/Message.vue'
import TimeDisplay from '@/components/misc/TimeDisplay.vue'
import XButton from '@/components/input/Button.vue'
import ImportEntryList from '@/components/manage/ImportEntryList.vue'

import {
	importStatusQuery,
	previewUserImportMutation,
	runUserImportMutation,
	uploadUserListMutation,
} from '@/client/queries/manage'
import type {ImportCounts, ImportPreview, ImportUploadResult} from '@/client/queries/manage'
import {error, success} from '@/message'
import {useTitle} from '@/composables/useTitle'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('manage.import.title'))

const status = useQuery(importStatusQuery())
const data = computed(() => status.data.value)

const uploadMutation = useMutation(uploadUserListMutation())
const previewMutation = useMutation(previewUserImportMutation())
const runMutation = useMutation(runUserImportMutation())

const fileInput = ref<HTMLInputElement | null>(null)
const chosenFile = ref<File | null>(null)
const uploaded = ref<ImportUploadResult | null>(null)
const previewResult = ref<ImportPreview | null>(null)
const previewed = computed(() => previewResult.value !== null)
const confirmRun = ref(false)

function formatBytes(n: number): string {
	if (n < 1024) return `${n} B`
	if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KiB`
	return `${(n / 1024 / 1024).toFixed(1)} MiB`
}

function summary(c: ImportCounts): string {
	return t('manage.import.summary', {
		created: c.created,
		updated: c.updated,
		reenabled: c.reenabled,
		disabled: c.disabled,
		unchanged: c.unchanged,
		exempt: c.exempt,
		failed: c.failed,
	})
}

function onFileChosen(event: Event) {
	chosenFile.value = (event.target as HTMLInputElement).files?.[0] ?? null
}

async function upload() {
	if (!chosenFile.value) return
	try {
		uploaded.value = await uploadMutation.mutateAsync(chosenFile.value)
		// A new file makes any earlier preview wrong.
		previewResult.value = null
		chosenFile.value = null
		if (fileInput.value) fileInput.value.value = ''
		success({message: t('manage.import.uploadDone')})
	} catch (e) {
		error(e)
	}
}

async function preview() {
	try {
		previewResult.value = await previewMutation.mutateAsync()
	} catch (e) {
		previewResult.value = null
		error(e)
	}
}

async function run() {
	try {
		await runMutation.mutateAsync()
		confirmRun.value = false
		// The plan the admin approved is spent: the next run needs a fresh preview.
		previewResult.value = null
		success({message: t('manage.import.started')})
	} catch (e) {
		confirmRun.value = false
		error(e)
	}
}
</script>

<style lang="scss" scoped>
.manage-import__meta {
	display: grid;
	grid-template-columns: auto 1fr;
	column-gap: 1rem;
	row-gap: 0.25rem;
	margin-block-end: 0.5rem;

	dt {
		font-weight: 600;
		color: var(--grey-700);
	}

	dd {
		margin: 0;
	}
}
</style>
