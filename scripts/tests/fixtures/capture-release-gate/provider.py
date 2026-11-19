"""Loopback model transport for the packaged native capture gate."""

from dataclasses import dataclass
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import itertools
import json
from pathlib import Path
import re
import shlex
from threading import Event, Thread
import time

TICKET = "TEST-960"
TITLE = "Native capture release gate"
MODEL = "gpt-5.4"
EDIT = "printf 'def answer():\\n    return 42\\n' > fixture.py"
VERIFY = "python3 -B -c 'from fixture import answer; assert answer() == 42'"
STAGES = r"AL96:(ROOT|IMPLEMENT|VERIFY|NESTED|RESUME|UNRELATED)"


def strings(value):
    if isinstance(value, str):
        yield value
    elif isinstance(value, list):
        for item in value:
            yield from strings(item)
    elif isinstance(value, dict):
        for item in value.values():
            yield from strings(item)


@dataclass(frozen=True)
class Action:
    step: str
    tool: str = ""
    args: dict | None = None


def request_stage(inputs):
    stage = ""
    for item in inputs:
        if item.get("role") == "user":
            matches = re.findall(STAGES, "\n".join(strings(item.get("content", ""))))
            if matches:
                stage = matches[-1]
    return stage


class FixtureModel:
    def __init__(self, manager, project):
        self.manager = str(manager)
        self.project = str(project)
        self.binding_engine = ""
        self.nested_supported = None
        self.release = Event()
        self.holding = Event()
        self.serial = itertools.count(1)
        self.errors = []

    def choose(self, request, codex):
        inputs = request.get("input" if codex else "messages", [])
        stage = request_stage(inputs)
        if not stage:
            raise ValueError("Native provider request did not identify its fixture role")

        prefix = "call_al96_" + stage.lower() + "_"
        prior = {
            value.removeprefix(prefix)
            for value in strings(inputs)
            if value.startswith(prefix) and " " not in value
        }

        def shell(step, command):
            if codex:
                return Action(step, "exec_command", {"cmd": command, "workdir": self.project, "max_output_tokens": 256})
            return Action(step, "bash", {"command": command, "workdir": self.project, "description": "Run fixture command"})

        def spawn(step, role):
            message = f"AL96:{role} Perform the assigned fixture work in the current repository."
            if codex:
                return Action(step, "spawn_agent", {"message": message, "fork_context": False})
            return Action(step, "task", {
                "description": f"Fixture {role.lower()}", "prompt": message, "subagent_type": "general",
            })

        def wait(step, spawn_step):
            call_id = prefix + spawn_step
            output = next(item["output"] for item in reversed(inputs)
                          if item.get("type") == "function_call_output" and item.get("call_id") == call_id)
            agent_id = re.search(r'"(?:agent_id|thread_id)"\s*:\s*"([^"]+)"', "\n".join(strings(output)))
            if not agent_id:
                raise ValueError(f"Expected native child ID in {spawn_step} output")
            return Action(step, "wait_agent", {"ids": [agent_id[1]], "timeout_ms": 30000})

        def check_wait(step):
            if not codex or step not in prior:
                return
            result = next(item["output"] for item in reversed(inputs)
                          if item.get("type") == "function_call_output" and item.get("call_id") == prefix + step)
            body = "\n".join(strings(result))
            if not re.search(r'"timed_out"\s*:\s*false', body) or not re.search(r'"(?:completed|Completed)"\s*:', body):
                raise ValueError(f"Native wait {step} did not return child completion")

        if stage == "UNRELATED":
            return shell("read", "cat fixture.py") if "read" not in prior else Action("done")
        if stage == "IMPLEMENT":
            if "edit" not in prior:
                return shell("edit", EDIT)
            self.nested_supported = codex or any(
                tool.get("function", {}).get("name") == "task" for tool in request.get("tools", [])
            )
            if not self.nested_supported:
                raise ValueError("Reviewed OpenCode version did not offer the nested task tool")
            if "nested" not in prior:
                return spawn("nested", "NESTED")
            if codex and "wait" not in prior:
                return wait("wait", "nested")
            check_wait("wait")
            return Action("done")
        if stage in ("VERIFY", "NESTED", "RESUME"):
            return shell("verify", VERIFY) if "verify" not in prior else Action("done")

        if "bind" not in prior:
            body = "\n".join(strings(request))
            if codex:
                pattern = r"execute:\s*(.+? capture submit-binding [^\n]*?)\s+--ticket-id <EXPLICIT_TICKET_ID>"
                found = re.search(pattern, body)
                if not found:
                    raise ValueError("Codex did not deliver its native binding challenge")
                command = shlex.split(found[1])
                if (not command or command[0] not in (self.manager, self.binding_engine)
                        or command[1:5] != ["capture", "submit-binding", "--repo", self.project]):
                    raise ValueError("Unexpected executable or project in Codex binding challenge")
                return shell("bind", shlex.join(command + [
                    "--ticket-id", TICKET, "--ticket-title", TITLE, "--ticket-type", "fix",
                ]))
            found = re.search(r"and request ([^\s.]+)", body)
            if not found:
                raise ValueError("OpenCode did not deliver its native binding challenge")
            return Action("bind", "agenticlens_bind", {
                "ticket": TICKET, "title": TITLE, "type": "fix", "request": found[1],
            })

        if "implementation" not in prior:
            return spawn("implementation", "IMPLEMENT")
        if codex and "waitimpl" not in prior:
            return wait("waitimpl", "implementation")
        check_wait("waitimpl")
        if "verification" not in prior:
            return spawn("verification", "VERIFY")
        if codex and "waitverify" not in prior:
            return wait("waitverify", "verification")
        check_wait("waitverify")
        return Action("hold")

    def begin(self):
        server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        server.model = self
        Thread(target=server.serve_forever, daemon=True).start()
        return server


def tool_definitions(tools):
    for tool in tools:
        if tool.get("type") == "namespace":
            for member in tool["tools"]:
                yield member, tool["name"]
        else:
            yield tool, ""


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.0"

    def log_message(self, *_args):
        pass

    def emit(self, value):
        self.wfile.write(("data: " + json.dumps(value) + "\n\n").encode())
        self.wfile.flush()

    def do_POST(self):
        model = self.server.model
        try:
            request = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
            codex = self.path.endswith("/responses")
            if not codex and not self.path.endswith("/chat/completions"):
                raise ValueError(f"Unexpected provider endpoint {self.path}")
            action = model.choose(request, codex)
            definitions = request.get("tools", [])
            namespace = ""
            if codex and action.tool:
                match = next(((tool, ns) for tool, ns in tool_definitions(definitions)
                              if tool.get("name") == action.tool), None)
                if match is None:
                    raise ValueError(f"Codex tool {action.tool!r} missing from provider request: "
                                     f"{[tool.get('name') for tool, _ in tool_definitions(definitions)]}")
                definition, namespace = match
                properties = definition.get("parameters", {}).get("properties", {})
                if action.tool == "spawn_agent" and "agent_type" in properties:
                    action.args["agent_type"] = "default"
                if action.tool == "wait_agent":
                    id_key = next((key for key in ("ids", "targets", "agent_ids") if key in properties), None)
                    if id_key is None:
                        raise ValueError(f"Unexpected Codex wait_agent schema keys: {list(properties)}")
                    action.args[id_key] = action.args.pop("ids")
            stage = request_stage(request.get("input" if codex else "messages", []))
            call_id = f"call_al96_{stage.lower()}_{action.step}"
        except (ValueError, KeyError, StopIteration) as exc:
            model.errors.append(str(exc))
            self.send_error(400, str(exc))
            return

        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Cache-Control", "no-cache")
        self.end_headers()
        try:
            if codex:
                self.stream_responses(request["model"], action, namespace, call_id)
            else:
                self.chat(request["model"], action, call_id)
        except (BrokenPipeError, ConnectionResetError):
            pass

    def hold(self):
        self.server.model.holding.set()
        if not self.server.model.release.wait(100):
            self.server.model.errors.append("Harness was held open but publication barrier was never released")

    def stream_responses(self, model, action, namespace, call_id):
        serial = next(self.server.model.serial)
        item_id = f"al96_item_{serial}"
        response = {"id": f"resp_al96_{serial}", "object": "response", "created_at": int(time.time()),
                    "status": "in_progress", "model": model, "output": []}
        self.emit({"type": "response.created", "response": response})
        if action.tool:
            item = {"id": item_id, "type": "function_call", "call_id": call_id,
                    "name": action.tool, "arguments": json.dumps(action.args), "status": "completed"}
            if namespace:
                item["namespace"] = namespace
            self.emit({"type": "response.output_item.added", "output_index": 0,
                       "item": dict(item, arguments="", status="in_progress")})
        else:
            item = {"id": item_id, "type": "message", "role": "assistant", "status": "completed",
                    "content": [{"type": "output_text", "text": "Fixture complete.", "annotations": []}]}
            self.emit({"type": "response.output_item.added", "output_index": 0,
                       "item": dict(item, content=[], status="in_progress")})
            self.emit({"type": "response.output_text.delta", "output_index": 0, "content_index": 0,
                       "item_id": item_id, "delta": "Fixture complete."})
            if action.step == "hold":
                self.hold()
        self.emit({"type": "response.output_item.done", "output_index": 0, "item": item})
        self.emit({"type": "response.completed", "response": dict(
            response, status="completed", output=[item],
            usage={"input_tokens": 23, "input_tokens_details": {"cached_tokens": 0},
                   "output_tokens": 11, "output_tokens_details": {"reasoning_tokens": 0}, "total_tokens": 34},
        )})

    def chat(self, model, action, call_id):
        def chunk(delta, reason=None, **extra):
            self.emit({"id": f"chatcmpl-{call_id}", "object": "chat.completion.chunk",
                       "created": int(time.time()), "model": model,
                       "choices": [{"index": 0, "delta": delta, "finish_reason": reason}], **extra})

        if action.tool:
            chunk({"role": "assistant", "tool_calls": [
                {"index": 0, "id": call_id, "type": "function",
                 "function": {"name": action.tool, "arguments": json.dumps(action.args)}},
            ]})
        else:
            chunk({"role": "assistant", "content": "Fixture complete."})
            if action.step == "hold":
                self.hold()
        chunk({}, "tool_calls" if action.tool else "stop",
              usage={"prompt_tokens": 23, "completion_tokens": 11, "total_tokens": 34})
        self.wfile.write(b"data: [DONE]\n\n")
