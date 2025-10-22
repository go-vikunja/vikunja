import { SSEServerTransport } from '@modelcontextprotocol/sdk/server/sse.js';
import type { Request, Response } from 'express';
import { logger } from '../../utils/logger.js';
import type { Authenticator } from '../../auth/authenticator.js';
import type { RateLimiter } from '../../ratelimit/limiter.js';

/**
 * Create and configure SSE transport for MCP
 */
export function createSSEServerTransport(
  endpoint: string,
  authenticator: Authenticator,
  rateLimiter: RateLimiter
) {
  return async (req: Request, res: Response): Promise<SSEServerTransport | null> => {
    const sessionId = `sse-${Date.now()}-${Math.random().toString(36).slice(2, 11)}`;

    try {
      // Extract and validate token from Authorization header or query parameter
      const authHeader = req.headers.authorization;
      const queryToken = req.query['token'] as string | undefined;
      
      let token: string | undefined;
      
      if (authHeader?.startsWith('Bearer ')) {
        token = authHeader.slice(7); // Remove 'Bearer ' prefix
      } else if (queryToken) {
        token = queryToken;
      }
      
      if (!token) {
        logger.warn('SSE connection rejected: missing token', {
          sessionId,
          ip: req.ip,
        });
        res.status(401).json({ error: 'Authorization required (Bearer token or ?token= parameter)' });
        return null;
      }

      // Authenticate token
      const userContext = await authenticator.validateToken(token);

      // Check rate limit (throws if exceeded)
      await rateLimiter.checkLimit(token);

      logger.info('SSE connection established', {
        sessionId,
        userId: userContext.userId,
        ip: req.ip,
      });

      // Create SSE transport
      const transport = new SSEServerTransport(endpoint, res);

      // Store user context for the connection
      (req as any).userId = userContext.userId;
      (req as any).sessionId = sessionId;

      // Handle connection close
      req.on('close', () => {
        logger.info('SSE connection closed', {
          sessionId,
          userId: userContext.userId,
        });
      });

      // Return transport so the server can be attached
      return transport;
    } catch (error) {
      logger.error('SSE transport error', {
        sessionId,
        error: error instanceof Error ? error.message : String(error),
      });

      if (!res.headersSent) {
        res.status(500).json({ error: 'Internal server error' });
      }

      return null;
    }
  };
}
