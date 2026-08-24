<template>
	<Dropdown
		trigger-icon="cog"
		:trigger-label="$t('planner.settings.title')"
	>
		<div class="planner-settings">
			<label class="setting">
				<span>{{ $t('planner.settings.dayStart') }}</span>
				<FormInput
					v-model="settings.dayStart"
					type="text"
					placeholder="HH:MM"
					pattern="[0-2][0-9]:[0-5][0-9]"
					inputmode="numeric"
				/>
			</label>
			<label class="setting">
				<span>{{ $t('planner.settings.dayEnd') }}</span>
				<FormInput
					v-model="settings.dayEnd"
					type="text"
					placeholder="HH:MM"
					pattern="[0-2][0-9]:[0-5][0-9]"
					inputmode="numeric"
				/>
			</label>
			<label class="setting">
				<span>{{ $t('planner.settings.defaultDuration') }}</span>
				<FormInput
					v-model.number="settings.defaultDurationMinutes"
					type="number"
					min="5"
					step="5"
				/>
			</label>
			<label class="setting">
				<span>{{ $t('planner.settings.slotDuration') }}</span>
				<FormInput
					v-model.number="settings.slotMinutes"
					type="number"
					min="5"
					step="5"
				/>
			</label>
			<FancyCheckbox v-model="settings.fullWeek">
				{{ $t('planner.settings.fullWeek') }}
			</FancyCheckbox>
			<label
				v-if="!settings.fullWeek"
				class="setting"
			>
				<span>{{ $t('planner.settings.daysToShow') }}</span>
				<FormInput
					v-model.number="settings.daysToShow"
					type="number"
					min="1"
					max="31"
					step="1"
				/>
			</label>
			<FancyCheckbox v-model="settings.showDone">
				{{ $t('planner.settings.showDone') }}
			</FancyCheckbox>
			<FancyCheckbox v-model="settings.showOverdue">
				{{ $t('planner.settings.showOverdue') }}
			</FancyCheckbox>
		</div>
	</Dropdown>
</template>

<script setup lang="ts">
import {watch} from 'vue'

import Dropdown from '@/components/misc/Dropdown.vue'
import FormInput from '@/components/input/FormInput.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import {useCalendarSettings} from './helpers/useCalendarSettings'

const {settings} = useCalendarSettings()

// Plain 24h "HH:MM" text fields (no native time picker, so no OS-locale AM/PM).
// Compare as minutes, not lexically, and only once both fields hold complete
// valid times — the watcher fires per keystroke, and clamping a half-typed
// value (e.g. '9') would hijack the input.
function timeToMinutes(time: string): number | null {
	const match = /^(\d{1,2}):(\d{2})$/.exec(time?.trim() ?? '')
	if (!match) {
		return null
	}
	const hours = Number(match[1])
	const minutes = Number(match[2])
	if (hours > 23 || minutes > 59) {
		return null
	}
	return hours * 60 + minutes
}

watch(() => settings.value.dayStart, start => {
	const startMinutes = timeToMinutes(start)
	const endMinutes = timeToMinutes(settings.value.dayEnd)
	if (startMinutes !== null && endMinutes !== null && startMinutes > endMinutes) {
		settings.value.dayStart = settings.value.dayEnd
	}
})
watch(() => settings.value.dayEnd, end => {
	const endMinutes = timeToMinutes(end)
	const startMinutes = timeToMinutes(settings.value.dayStart)
	if (endMinutes !== null && startMinutes !== null && endMinutes < startMinutes) {
		settings.value.dayEnd = settings.value.dayStart
	}
})
</script>

<style lang="scss" scoped>
.planner-settings {
	display: flex;
	flex-direction: column;
	gap: .6rem;
	min-inline-size: 14rem;
	padding: .25rem;
}

.setting {
	display: flex;
	flex-direction: column;
	gap: .2rem;
	font-size: .8rem;
}
</style>
