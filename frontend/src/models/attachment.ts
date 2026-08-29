import AbstractModel from './abstractModel'
import UserModel from './user'
import FileModel from './file'
import type { IUser } from '@/modelTypes/IUser'
import type { IFile } from '@/modelTypes/IFile'
import type { IAttachment } from '@/modelTypes/IAttachment'

export const SUPPORTED_IMAGE_SUFFIX = ['.jpeg', '.jpg', '.png', '.bmp', '.gif']
export const SUPPORTED_PDF_SUFFIX = ['.pdf']

export function canPreviewImage(attachment: IAttachment): boolean {
	const mime = attachment.file.mime.toLowerCase()
	// Gate on the sniffed mime, not just the extension; exclude svg since it can carry script.
	return SUPPORTED_IMAGE_SUFFIX.some((suffix) => attachment.file.name.toLowerCase().endsWith(suffix))
		&& mime.startsWith('image/')
		&& mime !== 'image/svg+xml'
}

export function canPreviewPdf(attachment: IAttachment): boolean {
	// Gate on the sniffed mime, not just the .pdf name: an HTML file named .pdf would otherwise run script in the same-origin preview iframe.
	return SUPPORTED_PDF_SUFFIX.some((suffix) => attachment.file.name.toLowerCase().endsWith(suffix))
		&& attachment.file.mime.toLowerCase() === 'application/pdf'
}

// No suffix allowlist, unlike images/pdfs: the blob is typed from the server's sniffed Content-Type, and an <audio> element neither parses HTML nor executes script.
export function canPreviewAudio(attachment: IAttachment): boolean {
	return attachment.file.mime.toLowerCase().startsWith('audio/')
}

// No suffix allowlist, for the same reason as audio: a <video> element neither parses HTML nor executes script, so the sniffed mime is the whole boundary.
export function canPreviewVideo(attachment: IAttachment): boolean {
	return attachment.file.mime.toLowerCase().startsWith('video/')
}

export type PreviewKind = 'image' | 'pdf' | 'video'

export function previewKind(attachment: IAttachment): PreviewKind | null {
	if (canPreviewImage(attachment)) {
		return 'image'
	}
	if (canPreviewPdf(attachment)) {
		return 'pdf'
	}
	if (canPreviewVideo(attachment)) {
		return 'video'
	}
	return null
}

export default class AttachmentModel extends AbstractModel<IAttachment> implements IAttachment {
	id = 0
	taskId = 0
	createdBy: IUser = UserModel
	file: IFile = FileModel
	created: Date = null

	constructor(data: Partial<IAttachment>) {
		super()
		this.assignData(data)

		this.createdBy = new UserModel(this.createdBy)
		this.file = new FileModel(this.file)
		this.created = new Date(this.created)
	}
}
