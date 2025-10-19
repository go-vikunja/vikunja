import { VikunjaClient } from '../vikunja/client.js';
import { UserContext } from '../auth/types.js';
import { VikunjaTask } from '../vikunja/types.js';
import { MCPResource, MCPResourceContents } from './projects.js';

/**
 * Task resources provider for MCP protocol
 */
export class TaskResources {
  constructor(private client: VikunjaClient) {}

  /**
   * List all tasks as MCP resources
   * @param userContext - Authenticated user context
   * @param page - Page number for pagination
   * @returns Array of MCP resources
   */
  async list(userContext: UserContext, page = 1): Promise<MCPResource[]> {
    this.client.setToken(userContext.token);
    const tasks = await this.client.get<VikunjaTask[]>('/api/v1/tasks/all', { page });

    return tasks.map((task) => ({
      uri: `vikunja://tasks/${task.id}`,
      name: task.title,
      description: task.description,
      mimeType: 'application/json',
    }));
  }

  /**
   * Read a single task by URI with expansion
   * @param uri - Resource URI (vikunja://tasks/{id})
   * @param userContext - Authenticated user context
   * @returns Resource contents with expanded data
   */
  async read(uri: string, userContext: UserContext): Promise<MCPResourceContents> {
    const taskId = this.parseTaskId(uri);
    this.client.setToken(userContext.token);
    
    // Get task with expansion for labels, assignees, etc.
    const task = await this.client.get<VikunjaTask>(`/api/v1/tasks/${taskId}`);

    return {
      contents: [
        {
          uri,
          mimeType: 'application/json',
          text: JSON.stringify(task, null, 2),
        },
      ],
    };
  }

  /**
   * Parse task ID from URI
   * @param uri - Resource URI
   * @returns Task ID
   */
  private parseTaskId(uri: string): number {
    const match = uri.match(/^vikunja:\/\/tasks\/(\d+)$/);
    if (!match || !match[1]) {
      throw new Error(`Invalid task URI: ${uri}`);
    }
    return parseInt(match[1], 10);
  }
}
