import { VikunjaClient } from '../vikunja/client.js';
import { UserContext } from '../auth/types.js';
import { VikunjaProject } from '../vikunja/types.js';

/**
 * MCP Resource representation
 */
export interface MCPResource {
  uri: string;
  name: string;
  description?: string;
  mimeType: string;
}

/**
 * MCP Resource contents
 */
export interface MCPResourceContents {
  contents: Array<{
    uri: string;
    mimeType: string;
    text: string;
  }>;
}

/**
 * Project resources provider for MCP protocol
 */
export class ProjectResources {
  constructor(private client: VikunjaClient) {}

  /**
   * List all projects as MCP resources
   * @param userContext - Authenticated user context
   * @param page - Page number for pagination
   * @returns Array of MCP resources
   */
  async list(userContext: UserContext, page = 1): Promise<MCPResource[]> {
    this.client.setToken(userContext.token);
    const projects = await this.client.get<VikunjaProject[]>('/api/v1/projects', { page });

    return projects.map((project) => ({
      uri: `vikunja://projects/${project.id}`,
      name: project.title,
      description: project.description,
      mimeType: 'application/json',
    }));
  }

  /**
   * Read a single project by URI
   * @param uri - Resource URI (vikunja://projects/{id})
   * @param userContext - Authenticated user context
   * @returns Resource contents
   */
  async read(uri: string, userContext: UserContext): Promise<MCPResourceContents> {
    const projectId = this.parseProjectId(uri);
    this.client.setToken(userContext.token);
    const project = await this.client.get<VikunjaProject>(`/api/v1/projects/${projectId}`);

    return {
      contents: [
        {
          uri,
          mimeType: 'application/json',
          text: JSON.stringify(project, null, 2),
        },
      ],
    };
  }

  /**
   * Parse project ID from URI
   * @param uri - Resource URI
   * @returns Project ID
   */
  private parseProjectId(uri: string): number {
    const match = uri.match(/^vikunja:\/\/projects\/(\d+)$/);
    if (!match || !match[1]) {
      throw new Error(`Invalid project URI: ${uri}`);
    }
    return parseInt(match[1], 10);
  }
}
