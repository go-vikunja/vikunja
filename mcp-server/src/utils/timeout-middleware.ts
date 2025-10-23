import type { Request, Response, NextFunction } from 'express';
import { logger } from '../utils/logger.js';

/**
 * Request timeout configuration
 */
export interface TimeoutConfig {
	/**
	 * Timeout in milliseconds (default: 30000 = 30 seconds)
	 */
	timeoutMs?: number;
}

/**
 * Express middleware for request timeout protection
 * 
 * Ensures long-running requests don't hang indefinitely, preventing
 * connection exhaustion and unresponsive server under slow client attacks.
 * 
 * @example
 * ```typescript
 * app.use('/mcp', timeoutMiddleware({ timeoutMs: 30000 }));
 * ```
 */
export function timeoutMiddleware(config?: TimeoutConfig) {
	const timeoutMs = config?.timeoutMs || 30000; // Default 30 seconds

	return (req: Request, res: Response, next: NextFunction) => {
		// Set timeout on the underlying socket
		if (req.socket) {
			req.socket.setTimeout(timeoutMs);
		}

		// Create timeout handler
		const timeoutId = setTimeout(() => {
			if (!res.headersSent) {
				logger.warn('Request timeout', {
					method: req.method,
					url: req.url,
					ip: req.ip,
					timeoutMs,
				});

				res.status(408).json({
					error: {
						code: 'REQUEST_TIMEOUT',
						message: `Request timed out after ${timeoutMs}ms`,
						data: {
							timeoutMs,
						},
					},
				});
			}
		}, timeoutMs);

		// Clear timeout on response finish
		res.on('finish', () => {
			clearTimeout(timeoutId);
		});

		// Clear timeout on response close (client disconnect)
		res.on('close', () => {
			clearTimeout(timeoutId);
		});

		next();
	};
}

/**
 * Middleware to halt processing if request has timed out
 * Use after timeout middleware to prevent processing timed-out requests
 */
export function haltOnTimeout(_req: Request, res: Response, next: NextFunction) {
	if (res.headersSent) {
		// Response already sent (possibly due to timeout)
		return;
	}
	next();
}
