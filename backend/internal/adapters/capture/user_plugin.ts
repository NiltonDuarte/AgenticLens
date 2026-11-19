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
    "ticket_id" in response && typeof response.ticket_id === "string") {
    const inherited = "binding" in response && response.binding === "source_parent";
    const outdated = "installation_id" in response && typeof response.installation_id === "string" &&
      response.installation_id !== "" &&
      (!("binding_installation_id" in response) || response.binding_installation_id !== response.installation_id);
    const incomplete = !("ticket_title" in response && response.ticket_title) ||
      !("ticket_type" in response && response.ticket_type);
    if (response.ticket_id === "" || (!inherited && (incomplete || outdated))) {
      const ticket = response.ticket_id === "" ? "the user-supplied active ticket" : `the recorded ticket ${response.ticket_id}`;
      const directory = outdated && response.ticket_id !== "" ? "" : ", optional existing destination directory";
      lines.push(`AgenticLens is collecting pending metrics for this OpenCode session. Call agenticlens_bind with ${ticket}, human-readable title, explicit conventional type${directory} and request ${response.request}. While required ticket information is unknown, capture remains pending.`);
    }
  }
  if ("destination_request" in response && typeof response.destination_request === "string" && response.destination_request &&
    "ticket_id" in response && typeof response.ticket_id === "string" && response.ticket_id) {
    lines.push(`AgenticLens has verified this session's ticket binding. If first publication still needs a destination and the directory is known, call agenticlens_destination with the bound ticket ${response.ticket_id}, exact existing directory and request ${response.destination_request}. The native worker validates the indicated directory before publication.`);
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
      description: "Bind this exact OpenCode session to the user's active ticket and optionally indicate an existing publication directory.",
      args: {
        ticket: tool.schema.string(),
        title: tool.schema.string(),
        type: tool.schema.string(),
        destination: tool.schema.string().optional(),
        request: tool.schema.string(),
      },
      execute: async ({ ticket, title, type, destination, request }, context) => notify("bind",
        { cwd: context.directory, session_id: context.sessionID, hook_event_name: "agent_binding" },
        ["--ticket-id", ticket, "--ticket-title", title, "--ticket-type", type, "--request", request,
          ...(destination === undefined ? [] : ["--destination-directory", destination])]),
    }),
    agenticlens_destination: tool({
      description: "Indicate the exact existing publication directory for this OpenCode session's bound ticket.",
      args: {
        ticket: tool.schema.string(),
        destination: tool.schema.string(),
        request: tool.schema.string(),
      },
      execute: async ({ ticket, destination, request }, context) => notify("bind",
        { cwd: context.directory, session_id: context.sessionID, hook_event_name: "agent_binding" },
        ["--destination-only", "--ticket-id", ticket, "--destination-directory", destination, "--request", request]),
    }),
  },
});
