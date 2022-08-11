import {string} from 'zod'
import isHexColor from 'validator/lib/isHexColor'

export const HexColorSchema = string().transform(
	(value) => {
		if (!value || value.startsWith('#')) {
			return value
		}
		return '#' + value
	}).refine(value => isHexColor(value))