<template>
	<div class="arrow-settings-wrapper">
		<button
			class="arrow-settings-toggle"
			:class="{ active: isOpen }"
			title="Dependency Arrow Settings"
			@click="isOpen = !isOpen"
		>
			<icon icon="sitemap" />
		</button>

		<div
			v-if="isOpen"
			class="arrow-settings-panel"
		>
			<div class="panel-header">
				<span class="panel-title">Dependency Arrows</span>
				<button class="panel-close" @click="isOpen = false">✕</button>
			</div>

			<!-- Master Enable/Disable -->
			<div class="setting-section master-toggle">
				<div class="setting-row toggle-row">
					<label>
						<input v-model="config.enabled" type="checkbox">
						<strong>Show dependency arrows</strong>
					</label>
				</div>
			</div>

			<!-- All other settings — disabled when arrows off -->
			<div :class="{ 'settings-disabled': !config.enabled }">

				<!-- Path Mode -->
				<div class="setting-section">
					<div class="section-label">Path Mode</div>
					<select v-model="config.pathMode" class="setting-select" :disabled="!config.enabled">
						<option value="bezier">Bezier (smooth)</option>
						<option value="stepped">Stepped (sharp)</option>
						<option value="stepRounded">Stepped + Rounded</option>
					</select>
				</div>

				<!-- Line Style -->
				<div class="setting-section">
					<div class="section-label">Line Style</div>
					<div class="setting-row">
						<label>Stroke</label>
						<input v-model.number="config.strokeWidth" type="range" min="0.5" max="4" step="0.25" :disabled="!config.enabled">
						<span class="setting-val">{{ config.strokeWidth }}</span>
					</div>
					<div class="setting-row">
						<label>Dash</label>
						<select v-model="config.dashArray" class="setting-select" :disabled="!config.enabled">
							<option value="4,2">Default</option>
							<option value="6,3">Long</option>
							<option value="2,2">Short</option>
							<option value="8,4">Wide</option>
							<option value="4,2,1,2">Dot-dash</option>
							<option value="none">Solid</option>
						</select>
					</div>
					<div class="setting-row">
						<label>Opacity</label>
						<input v-model.number="config.opacity" type="range" min="0.1" max="1" step="0.05" :disabled="!config.enabled">
						<span class="setting-val">{{ config.opacity }}</span>
					</div>
					<div class="setting-row">
						<label>Arrow</label>
						<input v-model.number="config.arrowSize" type="range" min="4" max="16" step="1" :disabled="!config.enabled">
						<span class="setting-val">{{ config.arrowSize }}</span>
					</div>
				</div>

				<!-- Bezier Controls -->
				<!-- Exit/Entry Edge — applies to ALL modes -->
				<div class="setting-section">
					<div class="section-label exit-label">Exit (Source)</div>
					<div class="setting-row">
						<label>Edge</label>
						<select v-model="config.exitDir" class="setting-select" :disabled="!config.enabled">
							<option value="right">Right →</option>
							<option value="bottom">Bottom ↓</option>
						</select>
					</div>
					<div class="setting-row">
						<label>Anchor</label>
						<input v-model.number="config.exitOffset" type="range" min="0" max="1" step="0.05" :disabled="!config.enabled">
						<span class="setting-val">{{ config.exitOffset }}</span>
					</div>
					<div class="setting-hint">{{ config.exitDir === 'bottom' ? '0 = left edge, 0.5 = center, 1 = right edge' : '0 = top, 0.5 = center, 1 = bottom' }}</div>
					<div class="setting-row" v-if="config.pathMode !== 'bezier'">
						<label>Length</label>
						<input v-model.number="config.exitLength" type="range" min="5" max="120" step="5" :disabled="!config.enabled">
						<span class="setting-val">{{ config.exitLength }}</span>
					</div>
				</div>

				<div class="setting-section">
					<div class="section-label entry-label">Entry (Target)</div>
					<div class="setting-row">
						<label>Edge</label>
						<select v-model="config.entryDir" class="setting-select" :disabled="!config.enabled">
							<option value="left">Left ←</option>
							<option value="top">Top ↑</option>
						</select>
					</div>
					<div class="setting-row">
						<label>Anchor</label>
						<input v-model.number="config.entryOffset" type="range" min="0" max="1" step="0.05" :disabled="!config.enabled">
						<span class="setting-val">{{ config.entryOffset }}</span>
					</div>
					<div class="setting-hint">{{ config.entryDir === 'left' ? '0 = top, 0.5 = center, 1 = bottom' : '0 = left edge, 0.5 = center, 1 = right edge' }}</div>
					<div class="setting-row" v-if="config.pathMode !== 'bezier'">
						<label>Length</label>
						<input v-model.number="config.entryLength" type="range" min="5" max="120" step="5" :disabled="!config.enabled">
						<span class="setting-val">{{ config.entryLength }}</span>
					</div>
				</div>

				<!-- Bezier-specific: curve control points -->
				<template v-if="config.pathMode === 'bezier'">
					<div class="setting-section">
						<div class="section-label cp1-label">CP1 — Source Arc</div>
						<div class="setting-row">
							<label>Horiz</label>
							<input v-model.number="config.cp1X" type="range" min="0.05" max="0.95" step="0.05" :disabled="!config.enabled">
							<span class="setting-val">{{ config.cp1X }}</span>
						</div>
						<div class="setting-row">
							<label>Vert ↕</label>
							<input v-model.number="config.cp1Y" type="range" min="-200" max="200" step="5" :disabled="!config.enabled">
							<span class="setting-val">{{ config.cp1Y }}</span>
						</div>
						<div class="setting-hint">Negative = UP above bars</div>
					</div>
					<div class="setting-section">
						<div class="section-label cp2-label">CP2 — Target Approach</div>
						<div class="setting-row">
							<label>Horiz</label>
							<input v-model.number="config.cp2X" type="range" min="0.05" max="0.95" step="0.05" :disabled="!config.enabled">
							<span class="setting-val">{{ config.cp2X }}</span>
						</div>
						<div class="setting-row">
							<label>Vert ↕</label>
							<input v-model.number="config.cp2Y" type="range" min="-200" max="200" step="5" :disabled="!config.enabled">
							<span class="setting-val">{{ config.cp2Y }}</span>
						</div>
						<div class="setting-hint">Negative = from above. Positive = from below</div>
					</div>
				</template>

				<!-- Stepped-specific: corner radius -->
				<div class="setting-section" v-if="config.pathMode === 'stepRounded'">
					<div class="section-label">Corners</div>
					<div class="setting-row">
						<label>Radius</label>
						<input v-model.number="config.cornerRadius" type="range" min="0" max="20" step="1" :disabled="!config.enabled">
						<span class="setting-val">{{ config.cornerRadius }}</span>
					</div>
				</div>

				<!-- Appearance -->
				<div class="setting-section">
					<div class="section-label">Appearance</div>
					<div class="setting-row">
						<label>Colors</label>
						<select v-model="config.palette" class="setting-select" :disabled="!config.enabled">
							<option value="multi">Multi-color</option>
							<option value="mono">Mono</option>
						</select>
					</div>
				</div>

				<!-- Extras -->
				<div class="setting-section">
					<div class="section-label">Extras</div>
					<div class="setting-row toggle-row">
						<label>
							<input v-model="config.showDots" type="checkbox" :disabled="!config.enabled">
							Source dots
						</label>
					</div>
					<div class="setting-row" v-if="config.showDots">
						<label>Dot size</label>
						<input v-model.number="config.dotRadius" type="range" min="1" max="6" step="0.5" :disabled="!config.enabled">
						<span class="setting-val">{{ config.dotRadius }}</span>
					</div>
					<div class="setting-row toggle-row">
						<label>
							<input v-model="config.showShadow" type="checkbox" :disabled="!config.enabled">
							Drop shadow
						</label>
					</div>
					<template v-if="config.showShadow">
						<div class="setting-row">
							<label>Width</label>
							<input v-model.number="config.shadowWidth" type="range" min="2" max="8" step="0.5" :disabled="!config.enabled">
							<span class="setting-val">{{ config.shadowWidth }}</span>
						</div>
						<div class="setting-row">
							<label>Opacity</label>
							<input v-model.number="config.shadowOpacity" type="range" min="0.05" max="0.5" step="0.05" :disabled="!config.enabled">
							<span class="setting-val">{{ config.shadowOpacity }}</span>
						</div>
					</template>
				</div>

			</div>

			<!-- Actions -->
			<div class="panel-actions">
				<button class="action-btn" @click="resetToDefaults">↩ Reset</button>
				<button class="action-btn" @click="copyConfig">📋 Copy</button>
				<button class="action-btn" @click="showImport = !showImport">📥 Import</button>
			</div>
			<div v-if="showImport" class="import-area">
				<textarea
					v-model="importJson"
					placeholder="Paste config JSON here..."
					rows="3"
				/>
				<button class="action-btn" @click="doImport">Apply</button>
			</div>
			<div v-if="statusMsg" class="status-msg">{{ statusMsg }}</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {useGanttArrowConfig} from '@/composables/useGanttArrowConfig'

const {config, resetToDefaults, importConfig, exportConfig} = useGanttArrowConfig()

const isOpen = ref(false)
const showImport = ref(false)
const importJson = ref('')
const statusMsg = ref('')

function copyConfig() {
	const json = exportConfig()
	navigator.clipboard.writeText(json).then(() => {
		statusMsg.value = '✓ Copied!'
		setTimeout(() => statusMsg.value = '', 2000)
	}).catch(() => {
		statusMsg.value = 'Copy failed'
	})
}

function doImport() {
	importConfig(importJson.value)
	showImport.value = false
	importJson.value = ''
	statusMsg.value = '✓ Imported!'
	setTimeout(() => statusMsg.value = '', 2000)
}
</script>

<style scoped lang="scss">
.arrow-settings-wrapper {
	position: relative;
	display: inline-block;
}

.arrow-settings-toggle {
	background: transparent;
	border: 1px solid var(--grey-300);
	border-radius: 4px;
	color: var(--grey-500);
	padding: 4px 8px;
	cursor: pointer;
	font-size: 0.8rem;
	display: flex;
	align-items: center;
	gap: 4px;
	transition: all 0.2s;

	&:hover, &.active {
		color: var(--primary);
		border-color: var(--primary);
		background: rgba(var(--primary-rgb, 93, 165, 218), 0.08);
	}
}

.arrow-settings-panel {
	position: absolute;
	top: 100%;
	right: 0;
	margin-top: 4px;
	width: 300px;
	max-height: 75vh;
	overflow-y: auto;
	background: var(--grey-100);
	border: 1px solid var(--grey-300);
	border-radius: 8px;
	padding: 0;
	z-index: 100;
	box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);

	.is-dark-mode & {
		background: rgba(25, 27, 38, 0.98);
		border-color: rgba(100, 120, 200, 0.3);
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
	}
}

.panel-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 8px 12px;
	border-bottom: 1px solid var(--grey-200);

	.is-dark-mode & {
		border-color: rgba(100, 120, 200, 0.15);
	}
}

.panel-title {
	font-size: 12px;
	font-weight: 600;
	color: var(--grey-800);

	.is-dark-mode & { color: #aab; }
}

.panel-close {
	background: none;
	border: none;
	cursor: pointer;
	color: var(--grey-500);
	font-size: 14px;
	padding: 0 4px;

	&:hover { color: var(--danger); }
}

.master-toggle {
	background: rgba(var(--primary-rgb, 93, 165, 218), 0.05);

	.is-dark-mode & {
		background: rgba(93, 165, 218, 0.08);
	}
}

.settings-disabled {
	opacity: 0.4;
	pointer-events: none;
}

.setting-section {
	padding: 8px 12px;
	border-bottom: 1px solid var(--grey-200);

	.is-dark-mode & {
		border-color: rgba(100, 120, 200, 0.1);
	}
}

.section-label {
	font-size: 9px;
	text-transform: uppercase;
	letter-spacing: 0.5px;
	color: var(--grey-500);
	margin-bottom: 4px;

	.is-dark-mode & { color: #668; }

	&.cp1-label { color: #e77; }
	&.cp2-label { color: #77e; }
	&.exit-label { color: #e77; }
	&.entry-label { color: #77e; }
}

.setting-row {
	display: flex;
	align-items: center;
	gap: 6px;
	margin-bottom: 4px;

	label {
		flex: 0 0 55px;
		font-size: 11px;
		color: var(--grey-600);

		.is-dark-mode & { color: #99a; }
	}

	input[type="range"] {
		flex: 1;
		height: 3px;
		accent-color: var(--primary);
	}

	&.toggle-row label {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 6px;
		cursor: pointer;
		font-size: 11px;

		input[type="checkbox"] {
			accent-color: var(--primary);
		}

		strong {
			font-weight: 600;
		}
	}
}

.setting-val {
	flex: 0 0 35px;
	text-align: right;
	font-size: 10px;
	font-family: monospace;
	color: var(--primary);
}

.setting-hint {
	font-size: 9px;
	color: var(--grey-400);
	margin: -2px 0 4px 0;

	.is-dark-mode & { color: #556; }
}

.setting-select {
	flex: 1;
	background: var(--grey-100);
	border: 1px solid var(--grey-300);
	border-radius: 4px;
	padding: 2px 4px;
	font-size: 11px;
	color: var(--grey-700);

	.is-dark-mode & {
		background: rgba(40, 42, 55, 0.9);
		border-color: rgba(100, 120, 200, 0.2);
		color: #ccd;
	}
}

.panel-actions {
	display: flex;
	gap: 4px;
	padding: 8px 12px;
}

.action-btn {
	flex: 1;
	padding: 4px 8px;
	border: 1px solid var(--grey-300);
	border-radius: 4px;
	background: var(--grey-100);
	color: var(--grey-600);
	font-size: 10px;
	cursor: pointer;
	text-align: center;

	&:hover {
		background: var(--grey-200);
		color: var(--grey-800);
	}

	.is-dark-mode & {
		background: rgba(50, 55, 80, 0.6);
		border-color: rgba(100, 120, 200, 0.2);
		color: #aab;

		&:hover {
			background: rgba(70, 75, 110, 0.7);
		}
	}
}

.import-area {
	padding: 0 12px 8px;

	textarea {
		width: 100%;
		font-family: monospace;
		font-size: 10px;
		border: 1px solid var(--grey-300);
		border-radius: 4px;
		padding: 4px;
		resize: vertical;
		margin-bottom: 4px;

		.is-dark-mode & {
			background: rgba(30, 32, 45, 0.8);
			border-color: rgba(100, 120, 200, 0.2);
			color: #ccd;
		}
	}
}

.status-msg {
	padding: 0 12px 8px;
	font-size: 10px;
	color: var(--success);
}
</style>
