import type { Plugin } from "@opencode-ai/plugin";
import { tool } from "@opencode-ai/plugin";
import { spawn } from "node:child_process";

type McpResult = {
  result?: unknown;
  error?: {
    message?: string;
  };
};

const MCP_COMMAND = process.env.NEABRAIN_MCP_COMMAND ?? "neabrain";
const MCP_ARGS = (process.env.NEABRAIN_MCP_ARGS?.trim() ?? "mcp")
  .split(" ")
  .filter(Boolean);

const SESSION_CONTEXT_QUERIES = new Map<string, string>();

const z = tool.schema;

function formatResult(result: unknown): string {
  if (typeof result === "string") {
    return result;
  }
  try {
    return JSON.stringify(result, null, 2);
  } catch {
    return String(result ?? "");
  }
}

function summarizeText(text: string, max = 500): string {
  const trimmed = text.trim().replace(/\s+/g, " ");
  if (trimmed.length <= max) {
    return trimmed;
  }
  return `${trimmed.slice(0, max - 3)}...`;
}

async function callMcpTool(name: string, args: Record<string, unknown>): Promise<unknown> {
  const payload = {
    jsonrpc: "2.0",
    id: "1",
    method: "tools/call",
    params: {
      name,
      arguments: args,
    },
  };

  const child = spawn(MCP_COMMAND, MCP_ARGS, {
    stdio: ["pipe", "pipe", "pipe"],
  });

  const stdout: string[] = [];
  const stderr: string[] = [];

  child.stdout.on("data", (chunk) => {
    stdout.push(String(chunk));
  });

  child.stderr.on("data", (chunk) => {
    stderr.push(String(chunk));
  });

  child.stdin.write(`${JSON.stringify(payload)}\n`);
  child.stdin.end();

  await new Promise<void>((resolve, reject) => {
    child.on("error", reject);
    child.on("close", () => resolve());
  });

  const output = stdout.join("").trim();
  if (!output) {
    const errorText = stderr.join("").trim();
    throw new Error(errorText || "neabrain mcp returned no output");
  }

  const line = output.split("\n").find((entry) => entry.trim().length > 0) ?? "";
  const parsed = JSON.parse(line) as McpResult;
  if (parsed.error?.message) {
    throw new Error(parsed.error.message);
  }

  return parsed.result ?? "";
}

async function safeMcpCall<T>(action: () => Promise<T>): Promise<T | null> {
  try {
    return await action();
  } catch {
    return null;
  }
}

async function getLastUserText(client: Parameters<Plugin>[0]["client"], sessionID: string): Promise<string> {
  const response = await client.session.messages({
    path: { id: sessionID },
  });

  for (let index = response.data.length - 1; index >= 0; index -= 1) {
    const entry = response.data[index];
    if (entry.info.role !== "user") {
      continue;
    }
    const text = entry.parts
      .filter((part) => part.type === "text")
      .map((part) => part.text)
      .join("\n");
    if (text.trim()) {
      return text;
    }
  }

  return "";
}

async function getLastAssistantText(
  client: Parameters<Plugin>[0]["client"],
  sessionID: string,
): Promise<string> {
  const response = await client.session.messages({
    path: { id: sessionID },
  });

  for (let index = response.data.length - 1; index >= 0; index -= 1) {
    const entry = response.data[index];
    if (entry.info.role !== "assistant") {
      continue;
    }
    const text = entry.parts
      .filter((part) => part.type === "text")
      .map((part) => part.text)
      .join("\n");
    if (text.trim()) {
      return text;
    }
  }

  return "";
}

function buildSessionSummary(sessionID: string, userText: string, assistantText: string): string {
  const lines = [`Session ${sessionID} compaction snapshot.`];
  if (userText) {
    lines.push(`Last user request: ${summarizeText(userText, 400)}`);
  }
  if (assistantText) {
    lines.push(`Last assistant response: ${summarizeText(assistantText, 400)}`);
  }
  return lines.join("\n");
}

function memoryInstructions(): string {
  return (
    "NeaBrain memory: use nbn_context to recall relevant observations and " +
    "nbn_session_summary to store short session summaries at key milestones."
  );
}

export const NeaBrainPlugin: Plugin = async ({ client }) => {
  const toolMap = {
    nbn_observation_create: {
      mcp: "observation.create",
      description: "Create a NeaBrain observation.",
      args: {
        content: z.string(),
        project: z.string().optional(),
        topic_key: z.string().optional(),
        tags: z.array(z.string()).optional(),
        source: z.string().optional(),
        metadata: z.record(z.any()).optional(),
        allow_duplicate: z.boolean().optional(),
      },
    },
    nbn_observation_read: {
      mcp: "observation.read",
      description: "Read a NeaBrain observation by id.",
      args: {
        id: z.string(),
        include_deleted: z.boolean().optional(),
      },
    },
    nbn_observation_update: {
      mcp: "observation.update",
      description: "Update a NeaBrain observation.",
      args: {
        id: z.string(),
        content: z.string().optional(),
        project: z.string().optional(),
        topic_key: z.string().optional(),
        tags: z.array(z.string()).optional(),
        source: z.string().optional(),
        metadata: z.record(z.any()).optional(),
      },
    },
    nbn_observation_list: {
      mcp: "observation.list",
      description: "List NeaBrain observations.",
      args: {
        project: z.string().optional(),
        topic_key: z.string().optional(),
        tags: z.array(z.string()).optional(),
        include_deleted: z.boolean().optional(),
      },
    },
    nbn_observation_delete: {
      mcp: "observation.delete",
      description: "Soft delete a NeaBrain observation.",
      args: {
        id: z.string(),
      },
    },
    nbn_search: {
      mcp: "search",
      description: "Search NeaBrain observations.",
      args: {
        query: z.string(),
        project: z.string().optional(),
        topic_key: z.string().optional(),
        tags: z.array(z.string()).optional(),
        include_deleted: z.boolean().optional(),
      },
    },
    nbn_topic_upsert: {
      mcp: "topic.upsert",
      description: "Upsert a NeaBrain topic.",
      args: {
        topic_key: z.string(),
        name: z.string().optional(),
        description: z.string().optional(),
        metadata: z.record(z.any()).optional(),
      },
    },
    nbn_session_open: {
      mcp: "session.open",
      description: "Open a NeaBrain session.",
      args: {
        disclosure_level: z.string(),
      },
    },
    nbn_session_resume: {
      mcp: "session.resume",
      description: "Resume a NeaBrain session.",
      args: {
        id: z.string(),
      },
    },
    nbn_session_update_disclosure: {
      mcp: "session.update_disclosure",
      description: "Update a NeaBrain session disclosure level.",
      args: {
        id: z.string(),
        disclosure_level: z.string(),
      },
    },
    nbn_config_show: {
      mcp: "config.show",
      description: "Show NeaBrain config.",
      args: {},
    },
  };

  const mappedTools = Object.fromEntries(
    Object.entries(toolMap).map(([toolName, config]) => {
      return [
        toolName,
        tool({
          description: config.description,
          args: config.args,
          async execute(args) {
            const result = await callMcpTool(config.mcp, args as Record<string, unknown>);
            return formatResult(result);
          },
        }),
      ];
    }),
  );

  const nbnSessionSummary = tool({
    description: "Store a concise session summary in NeaBrain.",
    args: {
      summary: z.string(),
      project: z.string().optional(),
      topic_key: z.string().optional(),
      tags: z.array(z.string()).optional(),
      metadata: z.record(z.any()).optional(),
    },
    async execute(args) {
      const payload = {
        content: args.summary,
        project: args.project ?? "",
        topic_key: args.topic_key ?? "",
        tags: args.tags ?? ["opencode", "session_summary"],
        source: "opencode",
        metadata: args.metadata ?? {},
        allow_duplicate: true,
      };
      const result = await callMcpTool("observation.create", payload);
      return formatResult(result);
    },
  });

  const nbnContext = tool({
    description: "Fetch NeaBrain context for a query.",
    args: {
      query: z.string(),
      project: z.string().optional(),
      topic_key: z.string().optional(),
      tags: z.array(z.string()).optional(),
      include_deleted: z.boolean().optional(),
    },
    async execute(args) {
      const payload = {
        query: args.query,
        project: args.project ?? "",
        topic_key: args.topic_key ?? "",
        tags: args.tags ?? [],
        include_deleted: args.include_deleted ?? false,
      };
      const result = await callMcpTool("search", payload);
      return formatResult(result);
    },
  });

  return {
    tool: {
      ...mappedTools,
      nbn_session_summary: nbnSessionSummary,
      nbn_context: nbnContext,
    },
    "experimental.chat.system.transform": async (_input, output) => {
      output.system.push(memoryInstructions());
    },
    "experimental.session.compacting": async (input, output) => {
      const userText = await safeMcpCall(() => getLastUserText(client, input.sessionID));
      const assistantText = await safeMcpCall(() => getLastAssistantText(client, input.sessionID));
      const summary = buildSessionSummary(
        input.sessionID,
        userText ?? "",
        assistantText ?? "",
      );

      await safeMcpCall(() =>
        nbnSessionSummary.execute(
          { summary, metadata: { session_id: input.sessionID } },
          {
            sessionID: input.sessionID,
            messageID: "",
            agent: "plugin",
            directory: "",
            worktree: "",
            abort: new AbortController().signal,
            metadata() {},
            async ask() {},
          },
        ),
      );

      const query = summarizeText(userText ?? "", 160) || "session context";
      SESSION_CONTEXT_QUERIES.set(input.sessionID, query);

      const contextText = await safeMcpCall(() =>
        nbnContext.execute(
          { query },
          {
            sessionID: input.sessionID,
            messageID: "",
            agent: "plugin",
            directory: "",
            worktree: "",
            abort: new AbortController().signal,
            metadata() {},
            async ask() {},
          },
        ),
      );

      if (contextText) {
        output.context.push(`NeaBrain context:\n${contextText}`);
      }
    },
    event: async ({ event }) => {
      if (event.type !== "session.compacted") {
        return;
      }
      const sessionID = event.properties.sessionID;
      const query = SESSION_CONTEXT_QUERIES.get(sessionID) ?? "session context";
      SESSION_CONTEXT_QUERIES.delete(sessionID);
      const contextText = await safeMcpCall(() =>
        nbnContext.execute(
          { query },
          {
            sessionID,
            messageID: "",
            agent: "plugin",
            directory: "",
            worktree: "",
            abort: new AbortController().signal,
            metadata() {},
            async ask() {},
          },
        ),
      );
      if (!contextText) {
        return;
      }

      await client.session.prompt({
        path: { id: sessionID },
        body: {
          noReply: true,
          parts: [
            {
              type: "text",
              text: `NeaBrain context:\n${contextText}`,
            },
          ],
        },
      });
    },
  };
};
