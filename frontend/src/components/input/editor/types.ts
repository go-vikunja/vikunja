export type UploadCallback = (files: File[] | FileList) => Promise<string[]>

export interface BottomAction {
    title: string
    action: () => void,
}
