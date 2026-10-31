"""Deterministic local OpenAI-compatible endpoint for AL-58 runtime evidence."""
from argparse import ArgumentParser
from http.server import BaseHTTPRequestHandler, HTTPServer
import json
import time


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def do_GET(self) -> None:
        if self.path.endswith("/models"):
            body = json.dumps(
                {
                    "object": "list",
                    "data": [
                        {
                            "id": "mock-model",
                            "object": "model",
                            "created": 0,
                            "owned_by": "al58",
                        }
                    ],
                }
            ).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            return
        self.send_error(404)

    def do_POST(self) -> None:
        request_length = int(self.headers.get("Content-Length", "0"))
        payload = json.loads(self.rfile.read(request_length) or b"{}")
        messages = payload.get("messages", [])
        has_tool_result = any(
            message.get("role") == "tool" or message.get("tool_call_id")
            for message in messages
        )
        if has_tool_result:
            delta = {"role": "assistant", "content": "AL-58 deterministic runtime reply."}
            finish_reason = "stop"
        else:
            argument_name = "filePath" if self.server.runtime == "v1" else "path"
            delta = {
                "role": "assistant",
                "tool_calls": [
                    {
                        "index": 0,
                        "id": "call_al58_read_1",
                        "type": "function",
                        "function": {
                            "name": "read",
                            "arguments": json.dumps({argument_name: "marker.txt"}),
                        },
                    }
                ],
            }
            finish_reason = "tool_calls"

        now = int(time.time())
        chunks = [
            {
                "id": "chatcmpl-al58",
                "object": "chat.completion.chunk",
                "created": now,
                "model": "mock-model",
                "choices": [{"index": 0, "delta": delta, "finish_reason": None}],
            },
            {
                "id": "chatcmpl-al58",
                "object": "chat.completion.chunk",
                "created": now,
                "model": "mock-model",
                "choices": [{"index": 0, "delta": {}, "finish_reason": finish_reason}],
                "usage": {"prompt_tokens": 23, "completion_tokens": 11, "total_tokens": 34},
            },
        ]
        body = b"".join(
            ("data: " + json.dumps(chunk) + "\n\n").encode() for chunk in chunks
        ) + b"data: [DONE]\n\n"
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Cache-Control", "no-cache")
        self.send_header("Connection", "close")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
        self.wfile.flush()

    def log_message(self, _format: str, *_args: object) -> None:
        return


def main() -> None:
    parser = ArgumentParser()
    parser.add_argument("--runtime", choices=("v1", "v2"), required=True)
    args = parser.parse_args()
    port = 8766 if args.runtime == "v1" else 8765
    server = HTTPServer(("127.0.0.1", port), Handler)
    server.runtime = args.runtime
    print(f"AL-58 mock provider listening on 127.0.0.1:{port}", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
