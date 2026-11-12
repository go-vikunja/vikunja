import draggable from 'zhyswan-vuedraggable'
import type {VNode} from 'vue'

type DraggableInstance = InstanceType<typeof draggable>

// Upstream omits slot declarations and the item type.
export function draggableFor<T>() {
	return draggable as {
		new(): Omit<DraggableInstance, '$props' | '$slots'> & {
			$props: Omit<DraggableInstance['$props'], 'list' | 'modelValue'> & {
				list?: T[],
				modelValue?: T[],
				'onUpdate:modelValue'?: (value: T[]) => void,
			},
			$slots: {
				item: (props: {element: T, index: number}) => VNode[],
				header?: () => VNode[],
				footer?: () => VNode[],
			},
		},
	}
}
