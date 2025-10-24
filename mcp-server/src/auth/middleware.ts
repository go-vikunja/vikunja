import type { Request, Response, NextFunction } from 'express';
import { TokenValidator } from './token-validator.js';
import type { UserContext } from './types.js';
import { AuthenticationError } from '../utils/errors.js';
import { logAuth } from '../utils/logger.js';

/**
 * Extended Request with user context
 */
export interface AuthenticatedRequest extends Request {
	userContext?: UserContext;
	token?: string;
}

/**
 * Singleton token validator instance
 */
const tokenValidator = new TokenValidator();

/**
 * Extract token from Bearer header
 */
function extractBearerToken(authHeader: string | undefined): string | null {
	if (!authHeader) {
		return null;
	}

	const parts = authHeader.split(' ');
	if (parts.length !== 2 || parts[0] !== 'Bearer') {
		return null;
	}

	return parts[1] ?? null;
}

/**
 * Express middleware for Bearer token authentication
 */
export async function authenticateBearer(
	req: AuthenticatedRequest,
	res: Response,
	next: NextFunction
): Promise<void> {
	try {
		const token = extractBearerToken(req.headers.authorization);

		if (!token) {
			throw new AuthenticationError('Missing or invalid Authorization header', {
				code: 'MISSING_TOKEN',
			});
		}

		// Validate token and get user context
		const userContext = await tokenValidator.validateToken(token);

		// Attach to request
		req.userContext = userContext;
		req.token = token;

		next();
	} catch (error) {
		if (error instanceof AuthenticationError) {
			res.status(401).json({
				error: {
					code: error.code,
					message: error.message,
					data: error.data,
				},
			});
		} else {
			logAuth('auth_failed', undefined, { error: (error as Error).message });
			res.status(500).json({
				error: {
					code: 'INTERNAL_ERROR',
					message: 'Authentication failed',
				},
			});
		}
	}
}

/**
 * Express middleware for query parameter token authentication
 * Used for EventSource connections which cannot send custom headers
 */
export async function authenticateQuery(
	req: AuthenticatedRequest,
	res: Response,
	next: NextFunction
): Promise<void> {
	try {
		const token = req.query['token'] as string | undefined;

		if (!token) {
			throw new AuthenticationError('Missing token query parameter', {
				code: 'MISSING_TOKEN',
			});
		}

		// Validate token and get user context
		const userContext = await tokenValidator.validateToken(token);

		// Attach to request
		req.userContext = userContext;
		req.token = token;

		next();
	} catch (error) {
		if (error instanceof AuthenticationError) {
			res.status(401).json({
				error: {
					code: error.code,
					message: error.message,
					data: error.data,
				},
			});
		} else {
			logAuth('auth_failed', undefined, { error: (error as Error).message });
			res.status(500).json({
				error: {
					code: 'INTERNAL_ERROR',
					message: 'Authentication failed',
				},
			});
		}
	}
}

/**
 * Get the singleton token validator instance
 */
export function getTokenValidator(): TokenValidator {
	return tokenValidator;
}
