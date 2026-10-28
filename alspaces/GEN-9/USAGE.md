# Using OpenCode and Codex with AgenticLens

OpenCode CLI **v1.19.5** is installed in this WSL user environment at `~/.opencode/bin/opencode`. Open a new WSL terminal, or run `source ~/.bashrc`, then check with:

```sh
opencode --version
```

If the shell cannot find it, run `~/.opencode/bin/opencode` directly.

## Sign OpenCode in with your ChatGPT plan

OpenCode 1.x documents OpenAI sign-in using ChatGPT Plus or Pro, with usage charged against the plan's included usage rather than OpenAI API billing. In the AgenticLens checkout:

```sh
cd /home/niltonduarte/.t3/worktrees/AgenticLens/t3-c935e1bc
opencode
```

At the OpenCode prompt, run `/connect`, choose **OpenAI**, and select **Sign in with ChatGPT** (or the available ChatGPT/Codex sign-in option). Complete the browser authorization with your ChatGPT account. Then run `/models` and choose an available OpenAI model. You can also start the login from a terminal with `opencode auth login openai`.

Do not paste an OpenAI API key if your goal is to use the ChatGPT subscription: API keys use separate API billing. OpenAI says eligible Go, Plus, and Pro subscribers can use plan usage in participating apps that support it; OpenCode's current v2 provider guide specifically lists Plus and Pro. Check the consent screen and your ChatGPT settings for the exact usage terms and limits shown for your account.

## Codex is already signed in here

The installed Codex CLI reports `Logged in using ChatGPT`. You can start it from the same checkout with:

```sh
cd /home/niltonduarte/.t3/worktrees/AgenticLens/t3-c935e1bc
codex
```

To inspect Codex login state later, use `codex login status`. OpenCode has its own authentication store, so complete its sign-in separately once.

## Working with both tools

- Both tools can open this same Git checkout and read the repository's `AGENTS.md` instructions.
- Use one tool at a time for edits. They share files, branch state, and the working tree; do not ask both to implement simultaneously.
- Give either tool the active ticket ID and ask it to follow the ticket workspace, iteration plan, contract, and review workflow. For this repository, every task needs a ticket; generic setup/documentation work uses a `GEN-*` ticket.
- Review `git status` and `git diff` before switching tools or committing. Keep each tool's sessions distinct so the work is easy to audit.
- To switch, leave the current session and launch the other CLI from the same repository directory. Neither tool needs a separate clone for sequential work.

Install OpenCode 1.x with the official installer (`curl -fsSL https://opencode.ai/install | bash`). OpenCode documentation: [installation](https://opencode.ai/docs/) and [providers](https://docs.opencode.ai/docs/providers/). OpenAI documentation: [ChatGPT plan usage in participating apps](https://help.openai.com/en/articles/20001542-using-your-chatgpt-plan-in-other-apps-and-sites).
