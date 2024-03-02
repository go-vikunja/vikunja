export interface IProvider {
	name: string;
	key: string;
	authUrl: string;
	clientId: string;
	logoutUrl: string;
	scope: string;
}
