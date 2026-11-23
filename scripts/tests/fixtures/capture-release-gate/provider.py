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
STAGES = (r"AL96:(ROOT|UNBOUND|BIND|DESTINATION|RETRY|IMPLEMENT|VERIFY|NESTED|RESUME|UNRELATED|COMPLETION|"
          r"DISCARD|DURING_DISCARD|RESTORE|DISCARD_BOUND|DURING_DISCARD_BOUND|RESTORE_BOUND)\b")
CONTROL_REQUEST = r"[0-9a-f]{64}\.(0|[1-9][0-9]*)\.[0-9a-f]{64}"
MANUAL_TICKET = "TEST-978"
FINAL_INPUT_TOKENS = 101
FINAL_OUTPUT_TOKENS = 47
WAIT_FOR_CHECKPOINT_RELEASE = """\
from pathlib import Path
import time

deadline = time.monotonic() + 22
while not Path('.fixture-release').exists():
    if time.monotonic() >= deadline:
        raise SystemExit('Timed out waiting for checkpoint observation')
    time.sleep(0.1)
"""


def manual_generate(manager, project, source, root):
    return [str(manager), "--al-offline", "--al-project", str(project), "generate", "--source", source,
            "--session-id", root, "--ticket-id", MANUAL_TICKET, "--ticket-type", "fix",
            "--ticket-title", "Manual completion preview", "--json-output",
            str(Path(project) / "selected" / "manual" / "agentic_metrics.json")]


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

    def usage(self):
        if self.step == "completion_reply":
            return FINAL_INPUT_TOKENS, FINAL_OUTPUT_TOKENS
        return 23, 11


def request_stage(inputs):
    stage = ""
    for item in inputs:
        if item.get("role") == "user":
            matches = re.findall(STAGES, "\n".join(strings(item.get("content", ""))))
            if matches:
                stage = matches[-1]
    return stage


class FixtureModel:
    def __init__(self, manager, project, destination):
        self.manager = str(manager)
        self.project = str(project)
        self.destination = str(destination)
        self.binding_engine = ""
        self.capture_control = None
        self.nested_supported = None
        self.release = Event()
        self.holding = Event()
        self.repair_release = Event()
        self.repair_holding = Event()
        self.serial = itertools.count(1)
        self.errors = []
        self.completion_root = None
        self.expected_completion_feedback = None
        self.feedback_in_model_input = Event()

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
        if stage == "COMPLETION":
            return self.completion_action(inputs, prior, prefix, codex, shell)
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
        if stage.startswith("DURING_DISCARD"):
            return shell("verify", VERIFY) if "verify" not in prior else Action("done")
        if stage == "RETRY":
            command = f"mkdir -- {shlex.quote(self.destination)} && {VERIFY}"
            return shell("create", command) if "create" not in prior else Action("hold_repair")

        def codex_submission(body, suffix):
            found = re.search(r"execute:\s*(.+? capture submit-binding [^\n]*?)\s+" + suffix, body)
            if not found:
                raise ValueError(f"Codex did not deliver the native challenge for {stage}")
            command = shlex.split(found[1])
            if (not command or command[0] not in (self.manager, self.binding_engine)
                    or command[1:5] != ["capture", "submit-binding", "--repo", self.project]):
                raise ValueError("Unexpected executable or project in Codex capture challenge")
            return command

        if stage in ("DISCARD", "RESTORE", "DISCARD_BOUND", "RESTORE_BOUND"):
            if "control" in prior:
                return Action("done")
            action = stage.split("_")[0].lower()
            command, request = self.native_control("\n".join(strings(request)), action, codex)
            if codex:
                return shell("control", shlex.join(command))
            return Action("control", "agenticlens_" + action,
                          {"capture_id": self.capture_control[1], "request": request})

        if stage == "DESTINATION":
            if "destination" in prior:
                return Action("done")
            body = "\n".join(strings(request))
            if codex:
                command = codex_submission(body, r"--destination-only\s+--ticket-id\b")
                return shell("destination", shlex.join(command + [
                    "--destination-only", "--ticket-id", TICKET, "--destination-directory", self.destination,
                ]))
            found = re.search(
                rf"call agenticlens_destination with the bound ticket {TICKET}, exact existing directory and request ([^\s.]+)",
                body, re.IGNORECASE,
            )
            if not found:
                raise ValueError("OpenCode did not deliver its native destination challenge")
            return Action("destination", "agenticlens_destination", {
                "ticket": TICKET, "destination": self.destination, "request": found[1],
            })

        if stage in ("ROOT", "BIND") and "bind" not in prior:
            body = "\n".join(strings(request))
            if codex:
                command = codex_submission(body, r"--ticket-id <EXPLICIT_TICKET_ID>")
                return shell("bind", shlex.join(command + [
                    "--ticket-id", TICKET, "--ticket-title", TITLE, "--ticket-type", "fix",
                ]))
            found = re.search(r"and request ([^\s.]+)", body)
            if not found:
                raise ValueError("OpenCode did not deliver its native binding challenge")
            return Action("bind", "agenticlens_bind", {
                "ticket": TICKET, "title": TITLE, "type": "fix", "request": found[1],
            })
        if stage == "BIND":
            return Action("done")

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

    def completion_action(self, inputs, prior, prefix, codex, shell):
        if "checkpoint" in prior:
            call_id = prefix + "checkpoint"
            if codex:
                result = next(item["output"] for item in inputs
                              if item.get("type") == "function_call_output" and item.get("call_id") == call_id)
            else:
                result = next(item["content"] for item in inputs
                              if item.get("role") == "tool" and item.get("tool_call_id") == call_id)
            if self.expected_completion_feedback and self.expected_completion_feedback in "\n".join(strings(result)):
                self.feedback_in_model_input.set()
            return Action("completion_reply")

        if codex:
            # exec_command yields at 30 seconds. These completed tool actions
            # all belong to the same native Codex turn.
            for index in range(7):
                step = f"work{index}"
                if step not in prior:
                    action = shell(step, "sleep 27")
                    action.args["yield_time_ms"] = 30000
                    return action

        source = "codex" if codex else "opencode"
        generate = shlex.join(manual_generate(self.manager, self.project, source, self.completion_root))
        command = " && ".join((
            EDIT, VERIFY, "git commit -qm 'Improve fixture answer' --only -- fixture.py",
            "python3 -B -c 'import time; print(time.time())' > .fixture-commit-time",
            f"{generate} > .fixture-generate-stdout 2> .fixture-generate-stderr",
            "cat .fixture-generate-stderr >&2",
            "touch .fixture-checkpoint-ready",
            f"python3 -B -c {shlex.quote(WAIT_FOR_CHECKPOINT_RELEASE)}",
        ))
        action = shell("checkpoint", command if codex else f"sleep 182 && {command}")
        if codex:
            action.args["yield_time_ms"] = 30000
        else:
            # OpenCode v1 associates this long running tool with a single
            # assistant message. Moving the preview to a later tool would
            # lose the exact still-open assistant interval.
            action.args["timeout"] = 220000
        return action

    def native_control(self, body, action, codex):
        root, key, revision, store = self.capture_control
        prefix = f"To {action} this capture when the user directs it, "
        if codex:
            commands = re.findall(re.escape(prefix) + r"execute: ([^\r\n]+)", body)
            for line in reversed(commands):
                command = shlex.split(line)
                if (len(command) != 15 or command[0] not in (self.manager, self.binding_engine)
                        or command[1:5] != ["capture", "submit-binding", "--repo", self.project]
                        or command[5:10] != ["--store", store, "--session-id", root, "--request"]
                        or command[11:] != ["--capture-id", key, "--capture-action", action]):
                    continue
                request = re.fullmatch(CONTROL_REQUEST, command[10])
                if request and int(request[1]) == revision:
                    return command, None
        else:
            pattern = re.escape(prefix + f"call agenticlens_{action} with capture ID {key} and request ")
            for request in reversed(list(re.finditer(pattern + "(" + CONTROL_REQUEST + r")(?![0-9A-Za-z.])", body))):
                if int(request[2]) == revision:
                    return None, request[1]
        raise ValueError(f"{action} did not receive a native control challenge for the expected capture and revision")

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
            if not codex and action.tool in ("agenticlens_discard", "agenticlens_restore") and not any(
                tool.get("function", {}).get("name") == action.tool for tool in definitions
            ):
                raise ValueError("OpenCode did not offer the expected native capture control tool")
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

    def hold(self, step):
        model = self.server.model
        entered, release = (model.repair_holding, model.repair_release) if step == "hold_repair" else (model.holding, model.release)
        entered.set()
        if not release.wait(100):
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
            if action.step in ("hold", "hold_repair"):
                self.hold(action.step)
        self.emit({"type": "response.output_item.done", "output_index": 0, "item": item})
        input_tokens, output_tokens = action.usage()
        self.emit({"type": "response.completed", "response": dict(
            response, status="completed", output=[item],
            usage={"input_tokens": input_tokens, "input_tokens_details": {"cached_tokens": 0},
                   "output_tokens": output_tokens, "output_tokens_details": {"reasoning_tokens": 0},
                   "total_tokens": input_tokens + output_tokens},
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
            if action.step in ("hold", "hold_repair"):
                self.hold(action.step)
        input_tokens, output_tokens = action.usage()
        chunk({}, "tool_calls" if action.tool else "stop",
              usage={"prompt_tokens": input_tokens, "completion_tokens": output_tokens,
                     "total_tokens": input_tokens + output_tokens})
        self.wfile.write(b"data: [DONE]\n\n")
