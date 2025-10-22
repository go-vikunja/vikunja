import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import {
  ListToolsRequestSchema,
  CallToolRequestSchema,
  type InitializeResult,
} from '@modelcontextprotocol/sdk/types.js';
import type { Server as HttpServer } from 'http';
import type { Authenticator } from './auth/authenticator.js';
import type { UserContext } from './auth/types.js';
import type { RateLimiter } from './ratelimit/limiter.js';
import type { VikunjaClient } from './vikunja/client.js';
import { ToolRegistry } from './tools/registry.js';
import { logger } from './utils/logger.js';
import { config } from './config/index.js';
import { createTransport } from './transport/factory.js';
import { createHttpTransportApp } from './transport/http.js';
import { SSEConnectionManager } from './transport/types.js';

/**
 * Initialize request parameters
 */
interface InitializeParams {
  protocolVersion: string;
  capabilities: Record<string, unknown>;
  clientInfo: {
    name: string;
    version: string;
  };
}

/**
 * MCP Server for Vikunja
 * Implements the Model Context Protocol for AI agent integration
 */
export class VikunjaMCPServer {
  private readonly server: Server;
  private readonly authenticator: Authenticator;
  private readonly userContexts: Map<string, UserContext>;
  private readonly toolRegistry: ToolRegistry;
  private readonly connectionManager: SSEConnectionManager;
  private httpServer: HttpServer | null = null;

  constructor(
    authenticator: Authenticator,
    rateLimiter: RateLimiter,
    vikunjaClient: VikunjaClient
  ) {
    this.authenticator = authenticator;
    this.userContexts = new Map();
    this.connectionManager = new SSEConnectionManager();

    // Initialize tool registry with client and rate limiter
    this.toolRegistry = new ToolRegistry(vikunjaClient, rateLimiter);
    this.toolRegistry.registerAllTools();

    // Initialize MCP server
    this.server = new Server(
      {
        name: 'vikunja-mcp',
        version: '1.0.0',
      },
      {
        capabilities: {
          resources: {},
          tools: {},
          prompts: {},
        },
      }
    );

    this.setupHandlers();
  }

  /**
   * Setup MCP protocol handlers
   */
  private setupHandlers(): void {
    // Register tools/list handler
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      logger.debug('Handling tools/list request');
      const tools = this.toolRegistry.getTools();
      return { tools };
    });

    // Register tools/call handler
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      logger.debug('Handling tools/call request', { tool: request.params.name });
      
      // Get user context from the request
      // In MCP, authentication is typically done during initialization
      // For now, we'll use a default context - this should be enhanced later
      const connectionId = 'default'; // TODO: Extract from request metadata
      const userContext = this.getUserContext(connectionId);
      
      if (!userContext) {
        logger.error('No user context found for connection', { connectionId });
        throw new Error('Unauthorized: No user context found');
      }

      try {
        const result = await this.toolRegistry.executeTool(
          request.params.name,
          request.params.arguments ?? {},
          userContext
        );
        
        return {
          content: [
            {
              type: 'text',
              text: JSON.stringify(result, null, 2),
            },
          ],
        };
      } catch (error) {
        logger.error('Tool execution failed', { 
          tool: request.params.name, 
          error: error instanceof Error ? error.message : String(error),
        });
        throw error;
      }
    });

    logger.info('MCP server handlers initialized');
  }

  /**
   * Handle MCP initialize request
   */
  async handleInitialize(request: InitializeParams): Promise<InitializeResult> {
    logger.info('Handling initialize request', {
      protocolVersion: request.protocolVersion,
      clientName: request.clientInfo.name,
      clientVersion: request.clientInfo.version,
    });

    return {
      protocolVersion: '2024-11-05',
      serverInfo: {
        name: 'vikunja-mcp',
        version: '1.0.0',
      },
      capabilities: {
        resources: {},
        tools: {},
        prompts: {},
      },
    };
  }

  /**
   * Handle MCP initialized notification
   */
  async handleInitialized(): Promise<void> {
    logger.info('Client initialized notification received');
  }

  /**
   * Authenticate a connection with Vikunja API token
   */
  async authenticateConnection(token: string): Promise<UserContext> {
    logger.info('Authenticating connection');
    const userContext = await this.authenticator.validateToken(token);
    logger.info('Connection authenticated', { userId: userContext.userId });
    return userContext;
  }

  /**
   * Store user context for a connection
   */
  setUserContext(connectionId: string, userContext: UserContext): void {
    this.userContexts.set(connectionId, userContext);
    logger.debug('User context stored', { connectionId, userId: userContext.userId });
  }

  /**
   * Get user context for a connection
   */
  getUserContext(connectionId: string): UserContext | undefined {
    return this.userContexts.get(connectionId);
  }

  /**
   * Remove user context for a connection
   */
  removeUserContext(connectionId: string): void {
    this.userContexts.delete(connectionId);
    logger.debug('User context removed', { connectionId });
  }

  /**
   * Start the MCP server with configured transport
   */
  async start(): Promise<void> {
    logger.info('Starting MCP server', { transportType: config.transportType });

    if (config.transportType === 'stdio') {
      // Stdio transport: direct connection
      const transport = createTransport(config);
      await this.server.connect(transport);
      logger.info('MCP server started with stdio transport');
    } else if (config.transportType === 'http') {
      // HTTP transport: start Express server
      this.startHttpTransport();
    } else {
      const exhaustiveCheck: never = config.transportType;
      throw new Error(`Unsupported transport type: ${String(exhaustiveCheck)}`);
    }
  }

  /**
   * Start HTTP transport server
   */
  private startHttpTransport(): void {
    if (typeof config.mcpPort !== 'number' || config.mcpPort === 0) {
      throw new Error('MCP_PORT is required for HTTP transport');
    }

    // Create Express app with SSE endpoint
    const app = createHttpTransportApp(
      this.authenticator,
      this.server,
      this.connectionManager
    );

    // Start HTTP server
    this.httpServer = app.listen(config.mcpPort, () => {
      logger.info('MCP HTTP transport server started', {
        port: config.mcpPort,
        endpoint: `/sse`,
      });
    });

    // Handle server errors
    this.httpServer.on('error', (error: Error) => {
      logger.error('HTTP server error', { error: error.message });
      throw error;
    });
  }

  /**
   * Stop the MCP server
   */
  async stop(): Promise<void> {
    logger.info('Stopping MCP server');

    // Close all SSE connections gracefully
    if (this.connectionManager !== null) {
      await this.connectionManager.closeAll();
      logger.info('All SSE connections closed');
    }

    // Close HTTP server if running
    if (this.httpServer !== null) {
      await new Promise<void>((resolve, reject) => {
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        this.httpServer!.close((err) => {
          if (err !== null && err !== undefined) {
            logger.error('Error closing HTTP server', { error: err.message });
            reject(err);
          } else {
            logger.info('HTTP server closed');
            resolve();
          }
        });
      });
      this.httpServer = null;
    }

    // Close MCP server
    await this.server.close();
    logger.info('MCP server stopped');
  }

  /**
   * Get the underlying MCP Server instance
   */
  getServer(): Server {
    return this.server;
  }
}
