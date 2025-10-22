import express, { type Request, type Response, type NextFunction, type Application } from 'express';
import { SSEServerTransport } from '@modelcontextprotocol/sdk/server/sse.js';
import type { Server } from '@modelcontextprotocol/sdk/server/index.js';
import type { Authenticator } from '../auth/authenticator.js';
import type { UserContext } from '../auth/types.js';
import { logger } from '../utils/logger.js';
import { v4 as uuidv4 } from 'uuid';
import { SSEConnectionManager } from './types.js';

/**
 * Extend Express Request to include user context
 */
// eslint-disable-next-line @typescript-eslint/no-namespace
declare global {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Express {
    interface Request {
      userContext?: UserContext;
    }
  }
}

/**
 * Authentication middleware for SSE endpoint
 * Validates Vikunja API token before establishing SSE connection
 */
export function createSSEAuthMiddleware(authenticator: Authenticator) {
  return async (req: Request, res: Response, next: NextFunction) => {
    try {
      // Extract token from Authorization header or query parameter
      const authHeader = req.headers.authorization;
      const tokenFromHeader = typeof authHeader === 'string' ? authHeader.replace('Bearer ', '') : undefined;
      const tokenFromQuery = req.query['token'] as string | undefined;
      const token = tokenFromHeader ?? tokenFromQuery;

      if (typeof token !== 'string' || token.length === 0) {
        logger.warn('SSE connection attempt without token', {
          ip: req.ip,
          userAgent: req.headers['user-agent'],
        });
        res.status(401).json({
          error: 'Unauthorized',
          message:
            'Missing authentication token. Provide token via Authorization header or ?token= query parameter.',
        });
        return;
      }

      // Validate token with existing Authenticator (includes 5-min cache)
      const userContext = await authenticator.validateToken(token);

      // Store user context in request for SSE handler
      req.userContext = userContext;

      logger.info('SSE connection authenticated', {
        userId: userContext.userId,
        username: userContext.username,
      });

      next();
    } catch (error) {
      logger.warn('SSE authentication failed', {
        error: error instanceof Error ? error.message : String(error),
        ip: req.ip,
      });

      res.status(401).json({
        error: 'Unauthorized',
        message: 'Invalid or expired authentication token',
      });
    }
  };
}

/**
 * SSE connection handler
 * Establishes SSE transport and connects MCP server
 */
export function createSSEConnectionHandler(
  mcpServer: Server,
  connectionManager: SSEConnectionManager
) {
  return async (req: Request, res: Response) => {
    const connectionId = uuidv4();
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    const userContext = req.userContext!; // Guaranteed by auth middleware

    logger.info('Establishing SSE connection', {
      connectionId,
      userId: userContext.userId,
      username: userContext.username,
    });

    try {
      // Create SSE transport for this connection
      const transport = new SSEServerTransport('/sse', res);

      // Connect MCP server to transport
      await mcpServer.connect(transport);

      logger.info('SSE connection established', {
        connectionId,
        userId: userContext.userId,
      });

      // Track connection for graceful shutdown
      connectionManager.add({
        id: connectionId,
        userContext,
        response: res,
        connectedAt: new Date(),
        lastActivityAt: new Date(),
        state: 'connected',
      });

      // Handle connection close
      res.on('close', () => {
        logger.info('SSE connection closed', {
          connectionId,
          userId: userContext.userId,
        });
        connectionManager.remove(connectionId);
      });
    } catch (error) {
      logger.error('Failed to establish SSE connection', {
        connectionId,
        userId: userContext.userId,
        error: error instanceof Error ? error.message : String(error),
      });

      if (!res.headersSent) {
        res.status(503).json({
          error: 'Service Unavailable',
          message: 'Failed to establish SSE connection. Please retry.',
        });
      }
    }
  };
}

/**
 * Create Express app for HTTP transport
 */
export function createHttpTransportApp(
  authenticator: Authenticator,
  mcpServer: Server,
  connectionManager: SSEConnectionManager
): Application {
  const app = express();

  // Parse JSON bodies (for future endpoints if needed)
  app.use(express.json());

  // SSE endpoint with authentication
  // eslint-disable-next-line @typescript-eslint/no-misused-promises
  app.post(
    '/sse',
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    createSSEAuthMiddleware(authenticator),
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    createSSEConnectionHandler(mcpServer, connectionManager)
  );

  logger.info('HTTP transport app created with /sse endpoint');

  return app;
}
