import {describe, it, expect, beforeEach} from 'vitest'
import {mount} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import DatepickerWithRange from './DatepickerWithRange.vue'
import en from '@/i18n/lang/en.json'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountPicker() {
    return mount(DatepickerWithRange, {
        props: {modelValue: {dateFrom: '', dateTo: ''}},
        global: {
            plugins: [i18n],
            stubs: ['RouterLink', 'Modal', 'XButton', 'BaseButton', 'Popup', 'flat-pickr'],
        },
    })
}

describe('DatepickerWithRange predefined ranges', () => {
    beforeEach(() => {
        setActivePinia(createPinia())
    })
    it('selects Last Week range', async () => {
        const wrapper = mountPicker()
        ;(wrapper.vm as any).setDateRange(['now/w-1w', 'now/w'])
        await wrapper.vm.$nextTick()
        const last = wrapper.emitted('update:modelValue')?.pop()?.[0]
        expect(last).toEqual({dateFrom: 'now/w-1w', dateTo: 'now/w'})
    })

    it('selects Last Month range', async () => {
        const wrapper = mountPicker()
        ;(wrapper.vm as any).setDateRange(['now/M-1M', 'now/M'])
        await wrapper.vm.$nextTick()
        const last = wrapper.emitted('update:modelValue')?.pop()?.[0]
        expect(last).toEqual({dateFrom: 'now/M-1M', dateTo: 'now/M'})
    })

    // A cleared range (the Custom option) comes back as null via v-model; the
    // modelValue watcher must coerce it, not call null.toISOString().
    it('accepts a null modelValue without crashing', async () => {
        const wrapper = mountPicker()
        await wrapper.setProps({modelValue: {dateFrom: 'now/w', dateTo: 'now/w+1w'}})
        await wrapper.vm.$nextTick()
        expect((wrapper.vm as any).from).toBe('now/w')

        await wrapper.setProps({modelValue: {dateFrom: null, dateTo: null}})
        await wrapper.vm.$nextTick()
        expect((wrapper.vm as any).from).toBe('')
        expect((wrapper.vm as any).to).toBe('')
    })
})
