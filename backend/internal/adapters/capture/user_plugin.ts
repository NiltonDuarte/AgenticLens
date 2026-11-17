import { execFile } from "node:child_process";
import { isAbsolute } from "node:path";
import { tool, type Plugin } from "@opencode-ai/plugin";

// The installer JSON-encodes the machine's native manager path.
const manager = __MANAGER__;

type SessionContext = {
  cwd: string;
  session_id: string;
  hook_event_name: string;
  activity?: string;
};

function notify(
  operation: "receive" | "request" | "bind",
  context: SessionContext,
  args: string[] = [],
): Promise<string> {
  if (!isAbsolute(context.cwd) || !context.session_id) {
    throw new Error("OpenCode did not supply an exact session ID and absolute directory");
  }
  return new Promise((resolve, reject) => {
    const child = execFile(manager, ["opencode-hook", operation, ...args],
      { timeout: 5000, maxBuffer: 65536 },
      (error, stdout, stderr) => {
        if (error) reject(new Error(stderr.trim() || error.message));
        else resolve(stdout);
      });
    child.stdin?.on("error", reject);
    child.stdin?.end(JSON.stringify(context));
  });
}

function describe(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function bindingFeedback(body: string): string[] {
  const response: unknown = JSON.parse(body);
  if (typeof response !== "object" || response === null) {
    throw new Error("Invalid native capture response");
  }
  const lines: string[] = [];
  if ("capture_feedback" in response && typeof response.capture_feedback === "string" && response.capture_feedback) {
    lines.push(response.capture_feedback);
  }
  if ("project_proof" in response && response.project_proof === "unverified") {
    lines.push("AgenticLens is waiting for exact source project evidence. Inspect capture diagnostics if this persists.");
  }
  if ("request" in response && typeof response.request === "string" &&
    "ticket_id" in response && (response.ticket_id === "" ||
    (!("binding" in response && response.binding === "source_parent") &&
    (!("ticket_title" in response) || response.ticket_title === "")))) {
    lines.push(`AgenticLens needs explicit ticket binding for this session. Call agenticlens_bind with the active ticket, human-readable title, optional conventional type and request ${response.request}. Use the ticket already supplied by the user. When the ticket is missing or ambiguous, ask the user and leave this session unassigned.`);
  }
  return lines;
}

export const AgenticLensCapture: Plugin = async ({ client, directory }) => ({
  event: async ({ event }) => {
    // OpenCode delivers the event hook inside its project's plugin instance.
    // Metadata-bearing events carry the session directory; status-only events
    // use the instance directory and native exact-session source verification.
    let context: SessionContext;
    switch (event.type) {
      case "session.created":
      case "session.updated":
        context = { cwd: event.properties.info.directory, session_id: event.properties.info.id, hook_event_name: event.type };
        break;
      case "session.status":
        context = { cwd: directory, session_id: event.properties.sessionID, hook_event_name: event.type, activity: event.properties.status.type };
        break;
      case "session.idle":
      case "session.error":
        if (!event.properties.sessionID) return;
        context = { cwd: directory, session_id: event.properties.sessionID, hook_event_name: event.type };
        break;
      default:
        return;
    }
    try {
      await notify("receive", context);
    } catch (error) {
      await client.app.log({ body: { service: "agenticlens-capture", level: "warn", message: describe(error) } });
    }
  },
  "experimental.chat.system.transform": async ({ sessionID }, output) => {
    if (!sessionID) return;
    try {
      const body = await notify("request", { cwd: directory, session_id: sessionID, hook_event_name: "agent_context_request" });
      output.system.push(...bindingFeedback(body));
    } catch (error) {
      output.system.push(`AgenticLens capture needs attention: ${describe(error)}`);
    }
  },
  tool: {
    agenticlens_bind: tool({
      description: "Bind this exact OpenCode session to the user's active ticket.",
      args: {
        ticket: tool.schema.string(),
        title: tool.schema.string(),
        type: tool.schema.string().optional(),
        request: tool.schema.string(),
      },
      execute: async ({ ticket, title, type, request }, context) => notify("bind",
        { cwd: context.directory, session_id: context.sessionID, hook_event_name: "agent_binding" },
        ["--ticket-id", ticket, "--ticket-title", title, "--request", request, ...(type ? ["--ticket-type", type] : [])]),
    }),
  },
});
