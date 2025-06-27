export interface ILoginCredentials {
	username: string | undefined
	password: string
	longToken: boolean
	totpPasscode?: string | undefined
}