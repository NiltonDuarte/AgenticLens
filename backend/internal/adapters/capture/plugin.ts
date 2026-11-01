import { execFile } from "node:child_process";
import type { Plugin } from "@opencode-ai/plugin";
import { tool } from "@opencode-ai/plugin";

// Values are JSON-encoded by the matching release binary, never evaluated as code.
const binary = __BINARY__;
const stateDir = __STATE__;
const sourceStore = __STORE__;
function call(args: string[], payload?: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const child = execFile(binary, ["capture", ...args, "--state-dir", stateDir],
      { timeout: 1500, maxBuffer: 65536, windowsHide: true },
      (error, stdout) => error ? reject(error) : resolve(stdout));
    if (payload !== undefined) child.stdin?.end(payload);
  });
}
const base = ["--source", "opencode", "--store", sourceStore];
export const AgenticLensCapture: Plugin = async ({ client }) => ({
  event: async ({ event }) => {
    if (!["session.created", "session.updated", "session.status", "session.idle", "session.error"].includes(event.type)) return;
    const props = event.properties;
    const id = "info" in props && typeof props.info === "object" && props.info !== null && "id" in props.info
      ? String(props.info.id) : "sessionID" in props && typeof props.sessionID === "string" ? props.sessionID : undefined;
    if (!id) return;
    try {
      await call(["receive", ...base], JSON.stringify({ session_id: id, hook_event_name: event.type }));
    } catch {
      await client.app.log({ body: { service: "agenticlens-capture", level: "warn", message: "Lifecycle capture failed; coding remains available. Run agentic-lens capture status." } });
    }
  },
  "experimental.chat.system.transform": async ({ sessionID }, output) => {
    if (!sessionID) return;
    try {
      const text = await call(["request", ...base, "--session-id", sessionID]);
      const response: unknown = JSON.parse(text);
      if (typeof response === "object" && response !== null && "request" in response && typeof response.request === "string" && "ticket_id" in response && response.ticket_id === "") {
        output.system.push(`AgenticLens needs explicit ticket binding for this session. Call agenticlens_bind with the ticket explicitly supplied by the user and request ${response.request}. Do not infer ticket ownership from directory, branch, role or timing. If the ticket is unknown, ask the user and leave this session unassigned. Source identity is supplied by the runtime; never invent parent/model/usage facts.`);
      }
    } catch {
      output.system.push("AgenticLens capture is degraded. Run agentic-lens capture status. Do not invent a ticket or claim complete capture.");
    }
  },
  tool: {
    agenticlens_bind: tool({
      description: "Respond to this session's AgenticLens binding request with an explicitly known ticket. Session identity is supplied by the runtime.",
      args: { ticket: tool.schema.string(), request: tool.schema.string() },
      execute: async ({ ticket, request }, context) => call(["bind", ...base, "--session-id", context.sessionID, "--ticket-id", ticket, "--request", request]),
    }),
  },
});
