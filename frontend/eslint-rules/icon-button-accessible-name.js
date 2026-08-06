/**
 * Flags icon-only buttons without an accessible name.
 *
 * BaseButton renders a bare <button>/<a> and does not enforce a name, so an
 * icon-only usage without aria-label is announced as just "button" (WCAG 1.1.1
 * / 4.1.2). Same for XButton's icon-only mode. Visible text, {{ }} content,
 * sr-only spans, aria-label(ledby) or title all count as a name; unknown child
 * components are assumed to render text so the rule only fires on clearly
 * icon-only content.
 */

const NAME_ATTRS = new Set(['aria-label', 'aria-labelledby', 'title'])
const ICON_ELEMENTS = new Set(['Icon', 'icon', 'svg', 'i'])
// Plain containers we can safely look through for text
const TRANSPARENT_ELEMENTS = new Set(['span', 'div', 'template'])

function attrName(attr) {
	if (!attr.directive) {
		return attr.key.rawName ?? attr.key.name
	}
	if (attr.key.name.name === 'bind' && attr.key.argument?.type === 'VIdentifier') {
		return attr.key.argument.rawName ?? attr.key.argument.name
	}
	return null
}

function hasNameAttr(element, extra = []) {
	return element.startTag.attributes.some(attr => {
		if (attr.directive && ['text', 'html'].includes(attr.key.name?.name)) {
			return true
		}
		const name = attrName(attr)
		if (name === null || !(NAME_ATTRS.has(name) || extra.includes(name))) {
			return false
		}
		// static attr must be non-empty; bound attrs are trusted
		return attr.directive || (attr.value != null && attr.value.value.trim() !== '')
	})
}

function getAttr(element, name) {
	return element.startTag.attributes.find(attr => attrName(attr) === name)
}

function contentProvidesName(element) {
	for (const child of element.children) {
		if (child.type === 'VText' && child.value.trim() !== '') {
			return true
		}
		if (child.type === 'VExpressionContainer') {
			return true
		}
		if (child.type === 'VElement') {
			if (child.rawName === 'slot' || child.rawName === 'component') {
				return true
			}
			if (hasNameAttr(child, ['alt'])) {
				return true
			}
			if (ICON_ELEMENTS.has(child.rawName)) {
				continue
			}
			if (TRANSPARENT_ELEMENTS.has(child.rawName)) {
				if (contentProvidesName(child)) {
					return true
				}
				continue
			}
			// Unknown component — assume it renders text
			return true
		}
	}
	return false
}

export default {
	meta: {
		type: 'problem',
		docs: {
			description: 'require an accessible name on icon-only BaseButton/XButton usages',
		},
		messages: {
			missingName: 'Icon-only <{{component}}> has no accessible name — screen readers announce it as just "button". Add aria-label (translated via $t) or visible/sr-only text.',
		},
		schema: [],
	},
	create(context) {
		const sourceCode = context.sourceCode ?? context.getSourceCode()
		const services = sourceCode.parserServices
		if (!services?.defineTemplateBodyVisitor) {
			return {}
		}

		function check(node, component) {
			if (hasNameAttr(node) || contentProvidesName(node)) {
				return
			}
			context.report({
				node: node.startTag,
				messageId: 'missingName',
				data: {component},
			})
		}

		return services.defineTemplateBodyVisitor({
			'VElement[rawName="BaseButton"]'(node) {
				check(node, 'BaseButton')
			},
			'VElement[rawName="XButton"]'(node) {
				// XButton is only icon-only when it has an icon prop and no slot content
				if (getAttr(node, 'icon') === undefined) {
					return
				}
				check(node, 'XButton')
			},
		})
	},
}
