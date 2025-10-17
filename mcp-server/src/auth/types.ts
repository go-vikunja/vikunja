/**
 * User context from authenticated token
 */
export interface UserContext {
  userId: number;
  username: string;
  email: string;
  token: string;
}
