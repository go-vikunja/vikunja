import AbstractModel from './abstractModel'
import UserModel from "./user";

export default class LabelModel extends AbstractModel {
	constructor(data) {
		super(data)
		// Set the default color
		if (this.hexColor === '') {
			this.hexColor = 'e8e8e8'
		}
		if (this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}
		this.textColor = this.hasDarkColor() ? '#4a4a4a' : '#e5e5e5'
		this.createdBy = new UserModel(this.createdBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
	
	defaults() {
		return {
			id: 0,
			title: '',
			hexColor: '',
			description: '',
			createdBy: UserModel,
			listId: 0,
			textColor: '',
			
			created: null,
			updated: null,
		}
	}
	
	hasDarkColor() {
		if (this.hexColor === '#') {
			return true // Defaults to dark
		}
		
		let rgb = parseInt(this.hexColor.substring(1, 7), 16);   // convert rrggbb to decimal
		let r = (rgb >> 16) & 0xff;  // extract red
		let g = (rgb >>  8) & 0xff;  // extract green
		let b = (rgb >>  0) & 0xff;  // extract blue
		
		// luma will be a value 0..255 where 0 indicates the darkest, and 255 the brightest
		let luma = 0.2126 * r + 0.7152 * g + 0.0722 * b; // per ITU-R BT.709
		return luma > 128
	}
}