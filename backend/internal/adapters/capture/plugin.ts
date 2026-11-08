import { execFile, spawn } from "node:child_process";
import type { Plugin } from "@opencode-ai/plugin";
import { tool } from "@opencode-ai/plugin";

// Values are JSON-encoded by the matching release binary, never evaluated as code.
const wrapper = __WRAPPER__;
const projectDir = __PROJECT__;
const stateDir = __STATE__;
const sourceStore = __STORE__;

// Windows owns the deadline, buffers and complete native manager/engine tree.
// Killing only the manager could leave its engine holding inherited pipes.
function callWindows(executable: string, args: string[], payload?: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const child = spawn(executable, args, { windowsHide: true });
    const output: Buffer[] = [];
    let outputBytes = 0;
    let errorBytes = 0;
    let settled = false;
    let cleanupStarted = false;
    let deadline: ReturnType<typeof setTimeout>;
    const cleanupTree = () => {
      if (cleanupStarted || child.pid === undefined || child.exitCode !== null || child.signalCode !== null) return;
      cleanupStarted = true;
      const taskkill = `${process.env.SystemRoot || "C:\\Windows"}\\System32\\taskkill.exe`;
      // child.pid is the actual Windows PID from this spawn, never payload data.
      const cleanup = spawn(taskkill, ["/PID", String(child.pid), "/T", "/F"], { windowsHide: true, stdio: "ignore" });
      const cleanupDeadline = setTimeout(() => {
        cleanup.kill();
        child.kill();
        console.error("AgenticLens Windows process-tree cleanup timed out; inspect capture diagnostics and running capture processes.");
      }, 1000);
      cleanup.on("error", () => {
        clearTimeout(cleanupDeadline);
        child.kill();
        console.error("AgenticLens Windows process-tree cleanup unavailable; inspect capture diagnostics and running capture processes.");
      });
      cleanup.once("exit", (code) => {
        clearTimeout(cleanupDeadline);
        if (code !== 0 && child.exitCode === null && child.signalCode === null) {
          child.kill();
          console.error("AgenticLens Windows process-tree cleanup failed; inspect capture diagnostics and running capture processes.");
        }
      });
    };
    const finish = (error: Error | null) => {
      if (settled) return;
      settled = true;
      clearTimeout(deadline);
      if (error) {
        cleanupTree();
        // Promise completion never waits for a descendant to release a pipe.
        child.stdin?.destroy();
        child.stdout?.destroy();
        child.stderr?.destroy();
        output.length = 0;
        reject(error);
      } else {
        resolve(Buffer.concat(output, outputBytes).toString("utf8"));
      }
    };
    deadline = setTimeout(() => finish(new Error("Windows capture handoff timed out; inspect capture diagnostics.")), 5000);
    child.on("error", (error) => finish(error));
    child.once("close", (code) => finish(code === 0 ? null : new Error("Windows capture handoff failed; inspect capture diagnostics.")));
    child.stdout?.on("data", (chunk: Buffer) => {
      if (settled) return;
      outputBytes += chunk.length;
      if (outputBytes > 65536) finish(new Error("Windows capture output exceeded its limit; inspect capture diagnostics."));
      else output.push(chunk);
    });
    child.stderr?.on("data", (chunk: Buffer) => {
      if (settled) return;
      errorBytes += chunk.length;
      if (errorBytes > 65536) finish(new Error("Windows capture diagnostics exceeded their limit; inspect capture diagnostics."));
    });
    child.stdin?.on("error", (error: Error) => finish(error));
    child.stdout?.on("error", (error: Error) => finish(error));
    child.stderr?.on("error", (error: Error) => finish(error));
    child.stdin?.end(payload);
  });
}

function call(args: string[], payload?: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const prefix = ["--al-offline", "--al-project", projectDir, "capture"];
    const windows = process.platform === "win32";
    const nativeArgs = [...prefix, ...args, "--state-dir", stateDir, "--repo", projectDir, "--wrapper", wrapper];
    if (windows) {
      callWindows(wrapper, nativeArgs, payload).then(resolve, reject);
      return;
    }
    const child = execFile(wrapper, nativeArgs,
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
      await call(["receive", ...base], JSON.stringify({ session_id: id, hook_event_name: event.type, activity: "status" in props && typeof props.status === "object" && props.status !== null && "type" in props.status ? String(props.status.type) : undefined }));
    } catch {
      await client.app.log({ body: { service: "agenticlens-capture", level: "warn", message: "Lifecycle capture failed; coding remains available. Run agentic-lens capture status." } });
    }
  },
  "experimental.chat.system.transform": async ({ sessionID }, output) => {
    if (!sessionID) return;
    try {
      const text = await call(["request", ...base, "--session-id", sessionID]);
      const response: unknown = JSON.parse(text);
      if (typeof response === "object" && response !== null && "capture_feedback" in response && typeof response.capture_feedback === "string" && response.capture_feedback.trim() !== "") {
        output.system.push(response.capture_feedback);
      }
      if (typeof response === "object" && response !== null && "project_proof" in response && response.project_proof === "unverified") {
        output.system.push("AgenticLens observed this session but source project metadata is not yet verified. Capture remains pending; inspect capture diagnostics for this exact project. Do not claim complete capture or treat this as a harness trust failure.");
      }
      if (typeof response === "object" && response !== null && "request" in response && typeof response.request === "string" && "ticket_id" in response && (response.ticket_id === "" || !("ticket_title" in response) || response.ticket_title === "" || ("installation_id" in response && typeof response.installation_id === "string" && response.installation_id !== "" && (!("binding_installation_id" in response) || response.binding_installation_id !== response.installation_id)))) {
        output.system.push(`AgenticLens needs explicit ticket binding for this session. Call agenticlens_bind with the ticket, human-readable title and optional conventional type explicitly supplied by the user and request ${response.request}. Do not infer ticket ownership from directory, branch, role or timing. If the ticket is unknown, ask the user and leave this session unassigned. Source identity is supplied by the runtime; never invent parent/model/usage facts.`);
      }
    } catch {
      output.system.push("AgenticLens capture is degraded. Run agentic-lens capture status. Do not invent a ticket or claim complete capture.");
    }
  },
  tool: {
    agenticlens_bind: tool({
      description: "Respond to this session's AgenticLens binding request with an explicitly known ticket. Session identity is supplied by the runtime.",
      args: { ticket: tool.schema.string(), title: tool.schema.string(), type: tool.schema.string().optional(), request: tool.schema.string() },
      execute: async ({ ticket, title, type, request }, context) => call(["bind", ...base, "--session-id", context.sessionID, "--ticket-id", ticket, "--ticket-title", title, "--request", request, ...(type ? ["--ticket-type", type] : [])]),
    }),
  },
});
