<template>
	<div class="content csv-migration">
		<h1>{{ $t('migrate.titleService', {name: 'CSV'}) }}</h1>
		<p>{{ $t('migrate.csv.description') }}</p>

		<!-- Step 1: File Upload -->
		<div
			v-if="step === 'upload'"
			class="upload-step"
		>
			<Message
				v-if="error"
				variant="danger"
				class="mbe-4"
			>
				{{ error }}
			</Message>
			<p>{{ $t('migrate.csv.uploadDescription') }}</p>
			<input
				ref="uploadInput"
				class="is-hidden"
				type="file"
				accept=".csv,.txt"
				@change="handleFileUpload"
			>
			<XButton
				:loading="isLoading"
				:disabled="isLoading || undefined"
				@click="uploadInput?.click()"
			>
				{{ $t('migrate.csv.selectFile') }}
			</XButton>
		</div>

		<!-- Step 2: Column Mapping -->
		<div
			v-else-if="step === 'mapping'"
			class="mapping-step"
		>
			<div class="mapping-header">
				<h2>{{ $t('migrate.csv.columnMapping') }}</h2>
				<p>{{ $t('migrate.csv.columnMappingDescription') }}</p>
			</div>

			<!-- Parsing Options -->
			<div class="parsing-options card">
				<h3>{{ $t('migrate.csv.parsingOptions') }}</h3>
				<div class="options-grid">
					<div class="option-group">
						<label for="delimiter">{{ $t('migrate.csv.delimiter') }}</label>
						<select
							id="delimiter"
							v-model="config.delimiter"
							@change="updatePreview"
						>
							<option
								v-for="delim in SUPPORTED_DELIMITERS"
								:key="delim.value"
								:value="delim.value"
							>
								{{ delim.label }}
							</option>
						</select>
					</div>
					<div class="option-group">
						<label for="dateFormat">{{ $t('migrate.csv.dateFormat') }}</label>
						<select
							id="dateFormat"
							v-model="config.date_format"
							@change="updatePreview"
						>
							<option
								v-for="format in SUPPORTED_DATE_FORMATS"
								:key="format.value"
								:value="format.value"
							>
								{{ format.label }}
							</option>
						</select>
					</div>
				</div>
			</div>

			<!-- Column Mappings -->
			<div class="column-mappings card">
				<h3>{{ $t('migrate.csv.mapColumns') }}</h3>
				<div class="mappings-grid">
					<div
						v-for="(mapping, index) in config.mapping"
						:key="index"
						class="mapping-row"
					>
						<div class="column-name">
							<strong>{{ mapping.column_name }}</strong>
							<span
								v-if="detectionResult && detectionResult.preview_rows[0]"
								class="preview-value"
							>
								{{ $t('migrate.csv.example') }}: {{ detectionResult.preview_rows[0][index] || '-' }}
							</span>
						</div>
						<select
							v-model="mapping.attribute"
							@change="updatePreview"
						>
							<option
								v-for="attr in TASK_ATTRIBUTES"
								:key="attr.value"
								:value="attr.value"
							>
								{{ $t('migrate.csv.attributes.' + attr.value) }}
							</option>
						</select>
					</div>
				</div>
			</div>

			<!-- Preview -->
			<div
				v-if="previewResult"
				class="preview-section card"
			>
				<h3>{{ $t('migrate.csv.preview') }}</h3>
				<p>{{ $t('migrate.csv.previewDescription', {count: previewResult.total_rows}) }}</p>

				<div
					v-if="previewResult.errors && previewResult.errors.length > 0"
					class="preview-errors"
				>
					<Message variant="warning">
						{{ $t('migrate.csv.previewErrors', {count: previewResult.error_count}) }}
					</Message>
				</div>

				<div class="preview-tasks">
					<div
						v-for="(task, index) in previewResult.tasks"
						:key="index"
						class="preview-task card"
					>
						<div class="task-title">
							<strong>{{ task.title || $t('migrate.csv.untitled') }}</strong>
							<span
								v-if="task.done"
								class="done-badge"
							>{{ $t('migrate.csv.completed') }}</span>
						</div>
						<div
							v-if="task.description"
							class="task-description"
						>
							{{ truncate(task.description, 100) }}
						</div>
						<div class="task-meta">
							<span v-if="task.due_date">
								{{ $t('migrate.csv.dueDate') }}: {{ task.due_date }}
							</span>
							<span v-if="task.priority > 0">
								{{ $t('migrate.csv.priority') }}: {{ task.priority }}
							</span>
							<span v-if="task.project">
								{{ $t('migrate.csv.project') }}: {{ task.project }}
							</span>
							<span v-if="task.labels && task.labels.length > 0">
								{{ $t('migrate.csv.labels') }}: {{ task.labels.join(', ') }}
							</span>
						</div>
					</div>
				</div>
			</div>

			<!-- Actions -->
			<div class="actions">
				<XButton
					variant="tertiary"
					@click="resetToUpload"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					:loading="isLoading"
					:disabled="!hasValidMapping || isLoading"
					@click="performImport"
				>
					{{ $t('migrate.csv.import') }}
				</XButton>
			</div>
		</div>

		<!-- Step 3: Success -->
		<div
			v-else-if="step === 'success'"
			class="success-step"
		>
			<Message class="mbe-4">
				{{ successMessage }}
			</Message>
			<XButton :to="{name: 'home'}">
				{{ $t('home.goToOverview') }}
			</XButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'

import Message from '@/components/misc/Message.vue'

import CSVMigrationService, {
	type DetectionResult,
	type ImportConfig,
	type PreviewResult,
	TASK_ATTRIBUTES,
	SUPPORTED_DELIMITERS,
	SUPPORTED_DATE_FORMATS,
} from '@/services/migrator/csvMigration'

import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'
import {getErrorText} from '@/message'

type Step = 'upload' | 'mapping' | 'success'

const {t} = useI18n({useScope: 'global'})

useTitle(() => t('migrate.titleService', {name: 'CSV'}))

const csvService = shallowReactive(new CSVMigrationService())

const step = ref<Step>('upload')
const error = ref('')
const successMessage = ref('')
const isLoading = ref(false)
const uploadInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const detectionResult = ref<DetectionResult | null>(null)
const previewResult = ref<PreviewResult | null>(null)

const config = ref<ImportConfig>({
	delimiter: ',',
	quote_char: '"',
	date_format: '2006-01-02',
	mapping: [],
})

const hasValidMapping = computed(() => {
	if (!config.value.mapping.length) return false
	// At least one column should be mapped to title
	return config.value.mapping.some(m => m.attribute === 'title')
})

function truncate(text: string, length: number): string {
	if (text.length <= length) return text
	return text.substring(0, length) + '...'
}

async function handleFileUpload() {
	const files = uploadInput.value?.files
	if (!files || files.length === 0) return

	selectedFile.value = files[0]
	error.value = ''
	isLoading.value = true

	try {
		const result = await csvService.detect(selectedFile.value)
		detectionResult.value = result

		// Apply detected values
		config.value = {
			delimiter: result.delimiter,
			quote_char: result.quote_char,
			date_format: result.date_format,
			mapping: result.suggested_mapping,
		}

		// Get initial preview
		await updatePreview()

		step.value = 'mapping'
	} catch (e) {
		error.value = getErrorText(e)
	} finally {
		isLoading.value = false
	}
}

async function updatePreview() {
	if (!selectedFile.value) return

	isLoading.value = true
	try {
		previewResult.value = await csvService.preview(selectedFile.value, config.value)
	} catch (e) {
		error.value = getErrorText(e)
	} finally {
		isLoading.value = false
	}
}

async function performImport() {
	if (!selectedFile.value || !hasValidMapping.value) return

	isLoading.value = true
	error.value = ''

	try {
		const result = await csvService.migrate(selectedFile.value, config.value)
		successMessage.value = result.message

		// Reload projects
		const projectStore = useProjectStore()
		await projectStore.loadAllProjects()

		step.value = 'success'
	} catch (e) {
		error.value = getErrorText(e)
	} finally {
		isLoading.value = false
	}
}

function resetToUpload() {
	step.value = 'upload'
	selectedFile.value = null
	detectionResult.value = null
	previewResult.value = null
	error.value = ''
	config.value = {
		delimiter: ',',
		quote_char: '"',
		date_format: '2006-01-02',
		mapping: [],
	}
}
</script>

<style lang="scss" scoped>
.csv-migration {
	max-inline-size: 900px;
	margin: 0 auto;
}

.card {
	background: var(--white);
	border-radius: var(--border-radius);
	padding: 1.5rem;
	margin-block-end: 1.5rem;
	box-shadow: var(--shadow-sm);
}

.mapping-header {
	margin-block-end: 1.5rem;

	h2 {
		margin-block-end: 0.5rem;
	}
}

.parsing-options {
	h3 {
		margin-block-end: 1rem;
	}
}

.options-grid {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
	gap: 1rem;
}

.option-group {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;

	label {
		font-weight: 500;
	}

	select {
		padding: 0.5rem;
		border: 1px solid var(--grey-300);
		border-radius: var(--border-radius);
		background: var(--white);
	}
}

.column-mappings {
	h3 {
		margin-block-end: 1rem;
	}
}

.mappings-grid {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.mapping-row {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: 1rem;
	align-items: center;
	padding: 0.75rem;
	background: var(--grey-100);
	border-radius: var(--border-radius);

	@media (width <= 600px) {
		grid-template-columns: 1fr;
	}
}

.column-name {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;

	.preview-value {
		font-size: 0.85rem;
		color: var(--grey-500);
	}
}

.preview-section {
	h3 {
		margin-block-end: 0.5rem;
	}
}

.preview-errors {
	margin-block: 1rem;
}

.preview-tasks {
	display: flex;
	flex-direction: column;
	gap: 1rem;
	margin-block-start: 1rem;
}

.preview-task {
	padding: 1rem;
	background: var(--grey-50);
}

.task-title {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	margin-block-end: 0.5rem;
}

.done-badge {
	font-size: 0.75rem;
	padding: 0.125rem 0.5rem;
	background: var(--success);
	color: var(--white);
	border-radius: var(--border-radius);
}

.task-description {
	color: var(--grey-600);
	margin-block-end: 0.5rem;
}

.task-meta {
	display: flex;
	flex-wrap: wrap;
	gap: 1rem;
	font-size: 0.85rem;
	color: var(--grey-500);
}

.actions {
	display: flex;
	gap: 1rem;
	justify-content: flex-end;
	margin-block-start: 1.5rem;
}

.success-step {
	text-align: center;
	padding: 2rem;
}
</style>
