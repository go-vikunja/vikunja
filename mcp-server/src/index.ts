import 'dotenv/config';
import express, { type Express } from 'express';
import { config } from './config/index.js';
import { logger } from './utils/logger.js';
import { RedisStorage } from './ratelimit/storage.js';
import { Authenticator } from './auth/authenticator.js';
import { RateLimiter } from './ratelimit/limiter.js';
import { VikunjaClient } from './vikunja/client.js';
import { VikunjaMCPServer } from './server.js';

/**
 * Application state
 */
interface AppState {
  redis: RedisStorage;
  vikunjaClient: VikunjaClient;
  authenticator: Authenticator;
  rateLimiter: RateLimiter;
  mcpServer: VikunjaMCPServer;
  httpServer: Express;
  startTime: number;
}

let appState: AppState | null = null;

/**
 * Initialize all components
 */
async function initializeApp(): Promise<AppState> {
  logger.info('Initializing Vikunja MCP Server');

  // Connect to Redis
  logger.info('Connecting to Redis', {
    host: config.redis.host,
    port: config.redis.port,
  });
  const redis = new RedisStorage();
  await redis.connect();

  // Create Vikunja API client
  logger.info('Creating Vikunja API client', {
    apiUrl: config.vikunjaApiUrl,
  });
  const vikunjaClient = new VikunjaClient();

  // Create authenticator
  const authenticator = new Authenticator();

  // Create rate limiter
  const rateLimiter = new RateLimiter(redis);

  // Create MCP server
  const mcpServer = new VikunjaMCPServer(authenticator, rateLimiter, vikunjaClient);

  // Create HTTP server for health checks
  const httpServer = express();
  httpServer.use(express.json());

  // Health check endpoint
  httpServer.get('/health', async (_req, res) => {
    const redisHealthy = await redis.isHealthy();
    const uptime = process.uptime();
    const status = redisHealthy ? 'ok' : 'degraded';

    res.status(redisHealthy ? 200 : 503).json({
      status,
      version: '1.0.0',
      uptime,
      redis: redisHealthy ? 'connected' : 'disconnected',
      timestamp: new Date().toISOString(),
    });
  });

  // Metrics endpoint (optional)
  httpServer.get('/metrics', (_req, res) => {
    res.json({
      requests: 0, // TODO: implement metrics tracking
      connections: 0, // TODO: implement connection tracking
      errors: 0, // TODO: implement error tracking
      uptime: process.uptime(),
    });
  });

  // Info endpoint
  httpServer.get('/info', (_req, res) => {
    res.json({
      name: 'vikunja-mcp',
      version: '1.0.0',
      protocol: 'MCP 2024-11-05',
      vikunjaApiUrl: config.vikunjaApiUrl,
    });
  });

  // SSE endpoint for MCP over HTTP
  // Store active SSE transports by session ID
  const activeSessions = new Map();
  
  // GET /sse - Client opens SSE connection to receive messages
  httpServer.get('/sse', async (req, res) => {
    try {
      // Extract and validate token from Authorization header or query parameter
      const authHeader = req.headers.authorization;
      const queryToken = req.query['token'] as string | undefined;
      
      let token: string | undefined;
      
      if (authHeader?.startsWith('Bearer ')) {
        token = authHeader.slice(7);
      } else if (queryToken) {
        token = queryToken;
      }
      
      if (!token) {
        logger.warn('SSE connection rejected: missing token');
        res.status(401).json({ error: 'Authorization required' });
        return;
      }

      // Authenticate token
      const userContext = await authenticator.validateToken(token);

      // Check rate limit
      await rateLimiter.checkLimit(token);

      logger.info('SSE connection established', { userId: userContext.userId });

      // Create SSE transport
      const transport = new (await import('@modelcontextprotocol/sdk/server/sse.js')).SSEServerTransport('/sse', res);
      
      // Store transport by session ID
      const sessionId = transport.sessionId;
      activeSessions.set(sessionId, { transport, userContext });
      
      // Clean up on close
      transport.onclose = () => {
        activeSessions.delete(sessionId);
        mcpServer.removeUserContext(sessionId);
        logger.info('SSE connection closed', { sessionId, userId: userContext.userId });
      };

      // Connect the MCP server to this transport
      await mcpServer.getServer().connect(transport);
      
      // Set user context for this session
      // Use 'http-session' as a shared context for HTTP connections
      // TODO: Implement per-session context when MCP SDK supports request metadata
      mcpServer.setUserContext('http-session', userContext);
    } catch (error) {
      logger.error('SSE GET error', { error: error instanceof Error ? error.message : String(error) });
      if (!res.headersSent) {
        res.status(500).json({ error: 'Internal server error' });
      }
    }
  });
  
  // POST /sse - Client sends messages to server
  httpServer.post('/sse', async (req, res) => {
    try {
      const sessionId = req.query['sessionId'] as string | undefined;
      
      if (!sessionId) {
        res.status(400).json({ error: 'sessionId required' });
        return;
      }

      const session = activeSessions.get(sessionId);
      if (!session) {
        res.status(404).json({ error: 'Session not found' });
        return;
      }

      // Handle the POST message with the existing transport
      // Pass req.body since express.json() middleware already parsed it
      await session.transport.handlePostMessage(req, res, req.body);
    } catch (error) {
      logger.error('SSE POST error', { error: error instanceof Error ? error.message : String(error) });
      if (!res.headersSent) {
        res.status(500).json({ error: 'Internal server error' });
      }
    }
  });

  // Start HTTP server
  const port = config.port;
  await new Promise<void>((resolve) => {
    httpServer.listen(port, () => {
      logger.info(`HTTP server listening on port ${port}`);
      logger.info(`SSE endpoint available at GET/POST http://localhost:${port}/sse`);
      resolve();
    });
  });

  // Start MCP server (stdio transport)
  logger.info('Starting MCP server');
  await mcpServer.start();

  return {
    redis,
    vikunjaClient,
    authenticator,
    rateLimiter,
    mcpServer,
    httpServer,
    startTime: Date.now(),
  };
}

/**
 * Graceful shutdown
 */
async function shutdown(): Promise<void> {
  logger.info('Shutting down Vikunja MCP Server');

  if (!appState) {
    return;
  }

  try {
    // Stop MCP server
    await appState.mcpServer.stop();

    // Disconnect from Redis
    await appState.redis.disconnect();

    logger.info('Shutdown complete');
  } catch (error) {
    logger.error('Error during shutdown', { error });
  }
}

/**
 * Error handler
 */
function handleError(error: Error): void {
  logger.error('Unhandled error', { error });
  process.exit(1);
}

// Handle shutdown signals
process.on('SIGTERM', async () => {
  await shutdown();
  process.exit(0);
});

process.on('SIGINT', async () => {
  await shutdown();
  process.exit(0);
});

// Handle uncaught errors
process.on('uncaughtException', handleError);
process.on('unhandledRejection', handleError);

// Start the application
initializeApp()
  .then((state) => {
    appState = state;
    logger.info('Vikunja MCP Server started successfully', {
      version: '1.0.0',
      port: config.port,
    });
  })
  .catch((error: Error) => {
    logger.error('Failed to start server', { error });
    process.exit(1);
  });
