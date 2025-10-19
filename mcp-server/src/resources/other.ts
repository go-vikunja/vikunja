import { VikunjaClient } from '../vikunja/client.js';
import { UserContext } from '../auth/types.js';
import { VikunjaLabel, VikunjaTeam, VikunjaUser } from '../vikunja/types.js';
import { MCPResource, MCPResourceContents } from './projects.js';

/**
 * Label resources provider for MCP protocol
 */
export class LabelResources {
  constructor(private client: VikunjaClient) {}

  async list(userContext: UserContext): Promise<MCPResource[]> {
    this.client.setToken(userContext.token);
    const labels = await this.client.get<VikunjaLabel[]>('/api/v1/labels');

    return labels.map((label) => ({
      uri: `vikunja://labels/${label.id}`,
      name: label.title,
      description: label.description,
      mimeType: 'application/json',
    }));
  }

  async read(uri: string, userContext: UserContext): Promise<MCPResourceContents> {
    const labelId = this.parseId(uri, 'labels');
    this.client.setToken(userContext.token);
    const label = await this.client.get<VikunjaLabel>(`/api/v1/labels/${labelId}`);

    return {
      contents: [
        {
          uri,
          mimeType: 'application/json',
          text: JSON.stringify(label, null, 2),
        },
      ],
    };
  }

  private parseId(uri: string, type: string): number {
    const match = uri.match(new RegExp(`^vikunja://${type}/(\\d+)$`));
    if (!match || !match[1]) {
      throw new Error(`Invalid ${type} URI: ${uri}`);
    }
    return parseInt(match[1], 10);
  }
}

/**
 * Team resources provider for MCP protocol
 */
export class TeamResources {
  constructor(private client: VikunjaClient) {}

  async list(userContext: UserContext): Promise<MCPResource[]> {
    this.client.setToken(userContext.token);
    const teams = await this.client.get<VikunjaTeam[]>('/api/v1/teams');

    return teams.map((team) => ({
      uri: `vikunja://teams/${team.id}`,
      name: team.name,
      description: team.description,
      mimeType: 'application/json',
    }));
  }

  async read(uri: string, userContext: UserContext): Promise<MCPResourceContents> {
    const teamId = this.parseId(uri, 'teams');
    this.client.setToken(userContext.token);
    const team = await this.client.get<VikunjaTeam>(`/api/v1/teams/${teamId}`);

    return {
      contents: [
        {
          uri,
          mimeType: 'application/json',
          text: JSON.stringify(team, null, 2),
        },
      ],
    };
  }

  private parseId(uri: string, type: string): number {
    const match = uri.match(new RegExp(`^vikunja://${type}/(\\d+)$`));
    if (!match || !match[1]) {
      throw new Error(`Invalid ${type} URI: ${uri}`);
    }
    return parseInt(match[1], 10);
  }
}

/**
 * User resources provider for MCP protocol
 * Note: Filtered by Vikunja permissions (users can only see users they have access to)
 */
export class UserResources {
  constructor(private client: VikunjaClient) {}

  async list(userContext: UserContext, query?: string): Promise<MCPResource[]> {
    this.client.setToken(userContext.token);
    const params = query ? { s: query } : {};
    const users = await this.client.get<VikunjaUser[]>('/api/v1/users', params);

    return users.map((user) => ({
      uri: `vikunja://users/${user.id}`,
      name: user.username,
      description: user.name || user.email,
      mimeType: 'application/json',
    }));
  }

  async read(uri: string, userContext: UserContext): Promise<MCPResourceContents> {
    const userId = this.parseId(uri, 'users');
    this.client.setToken(userContext.token);
    const user = await this.client.get<VikunjaUser>(`/api/v1/users/${userId}`);

    return {
      contents: [
        {
          uri,
          mimeType: 'application/json',
          text: JSON.stringify(user, null, 2),
        },
      ],
    };
  }

  private parseId(uri: string, type: string): number {
    const match = uri.match(new RegExp(`^vikunja://${type}/(\\d+)$`));
    if (!match || !match[1]) {
      throw new Error(`Invalid ${type} URI: ${uri}`);
    }
    return parseInt(match[1], 10);
  }
}
