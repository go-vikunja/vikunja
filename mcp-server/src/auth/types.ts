/**
 * User context from authenticated token
 * 
 * This is the unified UserContext type used throughout the application.
 * It combines authentication details with validation metadata.
 */
export interface UserContext {
  userId: number;
  username: string;
  email: string;
  token: string;
  permissions: string[];
  tokenScopes?: string[];
  validatedAt: Date;
}
