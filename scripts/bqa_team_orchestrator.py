#!/usr/bin/env python3
"""
BQA Team Orchestrator

Local orchestration layer for BQA-OS role prompts + GitHub Issues + Codex CLI.

Safety:
- Dry-run by default.
- Mutating operations require --execute.
- Long loops require bounded --max-cycles.

This script does not call private APIs directly. It shells out to:
- gh       for GitHub Issues/PRs
- codex    for local agent execution
- git      for branches/commits
"""
from __future__ import annotations

import argparse
import json
import os
import re
import shlex
import subprocess
import sys
import textwrap
import time
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Iterable, Optional

ROOT = Path.cwd()
TEAM_DIR = ROOT / ".bqa-team"
ROLES_DIR = TEAM_DIR / "roles"
BACKLOG_DIR = TEAM_DIR / "backlog"
GENERATED_DIR = TEAM_DIR / "generated"
ISSUES_DIR = GENERATED_DIR / "issues"
PROMPTS_DIR = GENERATED_DIR / "prompts"
STATE_FILE = TEAM_DIR / "state.json"

LABELS = {
    "bqa:business": "Business-originated task",
    "bqa:needs-arch": "Needs technical architecture review",
    "bqa:arch-approved": "Approved by technical architect",
    "bqa:ready-dev": "Ready for implementation",
    "bqa:in-dev": "Currently being implemented",
    "bqa:ready-qa": "Ready for QA verification",
    "bqa:qa-failed": "QA found issues",
    "bqa:bug": "Bug found by QA or business review",
    "bqa:ready-business": "Ready for business acceptance",
    "bqa:business-approved": "Accepted by business owner",
    "bqa:done": "Done",
    "bqa:static-site": "Static web application work",
    "bqa:game-ui": "Game-style team visualization work",
    "bqa:codex-team": "Codex team automation work",
}

ROLE_FILES = {
    "business": "Founder_Product_Sales_Implementation.md",
    "architect": "BQA_OS_Tech_Lead_Architect.md",
    "developer": "BQA_OS_Go_CLI_Implementer.md",
    "qa": "BQA_OS_QA_Test_Engineer.md",
    "designer": "Designer_Frontend.md",
    "devroom": "BQA_OS_Dev_Room.md",
}

ISSUE_TEMPLATE = """## Context

{context}

## Goal

{goal}

## Scope

### Create/change

{create_change}

### Do not touch

{do_not_touch}

## Architecture

Follow:

core use case
↓
port interface
↓
adapter implementation
↓
CLI wiring

Business logic must not live directly in Cobra commands.

## Behavior

{behavior}

## Acceptance criteria

{acceptance}

## Manual verification

```bash
{verification}
```

## Role routing

- Business owner: validates value and scope.
- Technical architect: validates architecture before development.
- Developer: implements only after architecture approval.
- QA: verifies and creates bug issues if acceptance criteria fail.
- Business owner: performs final acceptance.
"""


def now() -> str:
    return datetime.now(timezone.utc).isoformat()


def log(msg: str) -> None:
    print(f"[{datetime.now().strftime('%H:%M:%S')}] {msg}")


def run(cmd: list[str], *, execute: bool, cwd: Path = ROOT, capture: bool = False, check: bool = True) -> subprocess.CompletedProcess[str]:
    printable = " ".join(shlex.quote(c) for c in cmd)
    if not execute:
        print(f"DRY-RUN: {printable}")
        return subprocess.CompletedProcess(cmd, 0, stdout="", stderr="")
    log(f"RUN: {printable}")
    return subprocess.run(cmd, cwd=str(cwd), text=True, capture_output=capture, check=check)


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def write(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")


def slugify(text: str, max_len: int = 56) -> str:
    text = text.lower()
    text = re.sub(r"[^a-z0-9]+", "-", text)
    text = text.strip("-") or "task"
    return text[:max_len].strip("-")


def load_role(role: str) -> str:
    path = ROLES_DIR / ROLE_FILES[role]
    if not path.exists():
        raise SystemExit(f"Missing role prompt: {path}")
    return read(path)


def load_state() -> dict:
    if STATE_FILE.exists():
        return json.loads(read(STATE_FILE))
    return {"processed_backlog": {}, "created_issues": {}, "runs": []}


def save_state(state: dict) -> None:
    write(STATE_FILE, json.dumps(state, indent=2, ensure_ascii=False) + "\n")


def require_tools(names: Iterable[str], execute: bool) -> None:
    for name in names:
        result = subprocess.run(["bash", "-lc", f"command -v {shlex.quote(name)}"], text=True, capture_output=True)
        if result.returncode != 0:
            msg = f"Required tool not found: {name}"
            if execute:
                raise SystemExit(msg)
            log(f"WARNING: {msg}")


def cmd_init(args: argparse.Namespace) -> None:
    TEAM_DIR.mkdir(exist_ok=True)
    for d in [ROLES_DIR, BACKLOG_DIR, ISSUES_DIR, PROMPTS_DIR]:
        d.mkdir(parents=True, exist_ok=True)
    if not STATE_FILE.exists():
        save_state(load_state())
    log(f"Initialized {TEAM_DIR}")
    log("Put business task markdown files into .bqa-team/backlog/")


def cmd_seed(args: argparse.Namespace) -> None:
    BACKLOG_DIR.mkdir(parents=True, exist_ok=True)
    seeds = {
        "001_static_site_app.md": """# Business Task: Static BQA Web App MVP

Create a static HTML/JS application for BQA-OS where a user can upload a specially marked session archive and receive a generated output archive containing agents, workflows, specs, instructions, and recommendations.

Constraints:
- static site only: HTML/CSS/JavaScript;
- local-first by default;
- no private data uploaded externally;
- output downloadable as zip;
- include synthetic sample input only.
""",
        "002_agent_game_visualization.md": """# Business Task: Agent Team Visualization / Warcraft-style Map MVP

Create a lightweight static dashboard that represents BQA agents as units on a board/map and shows task flow from business idea to architecture review, development, QA, bug fixing, business acceptance, and done.
""",
        "003_codex_team_pipeline.md": """# Business Task: Codex Team Pipeline MVP

Create a scripted workflow where business tasks are transformed into GitHub issues, routed through architecture review, implemented by role-specific Codex agents, checked by QA, and finally sent to business acceptance.
""",
    }
    for name, body in seeds.items():
        path = BACKLOG_DIR / name
        if not path.exists():
            write(path, body)
            log(f"Seeded {path}")


def architect_prompt(task_name: str, task_body: str, repo: str) -> str:
    architect = load_role("architect")
    devroom = load_role("devroom") if (ROLES_DIR / ROLE_FILES["devroom"]).exists() else ""
    return f"""
{architect}

---

Additional Dev Room context:

{devroom[:8000]}

---

You are transforming a business task into one or more GitHub-ready issue specs for repository `{repo}`.

Business task file: {task_name}

Business task:

{task_body}

Rules:
- Every business task must pass through technical architecture review before development.
- Keep issues small and implementation-ready.
- Follow the BQA-OS issue template.
- Preserve hexagonal architecture.
- No private data, real session logs, or secrets.
- For static site tasks, prefer plain HTML/CSS/JS MVP unless architecture says otherwise.
- For game visualization, start with simple board/map UI, not a heavy game engine.

Return markdown with this exact structure for each issue:

---ISSUE---
TITLE: <short title>
LABELS: bqa:arch-approved,bqa:ready-dev,<domain labels>
BODY:
<full GitHub issue body using the project template>
---END---
""".strip()


def cmd_architect(args: argparse.Namespace) -> None:
    require_tools(["codex"], args.execute)
    state = load_state()
    BACKLOG_DIR.mkdir(parents=True, exist_ok=True)
    tasks = sorted(BACKLOG_DIR.glob("*.md"))
    if not tasks:
        log("No backlog markdown files found in .bqa-team/backlog/")
        return
    for path in tasks:
        if state["processed_backlog"].get(path.name) and not args.force:
            log(f"Skip already architected: {path.name}")
            continue
        body = read(path)
        prompt = architect_prompt(path.name, body, args.repo)
        prompt_path = PROMPTS_DIR / f"architect_{path.stem}.md"
        output_path = ISSUES_DIR / f"{path.stem}.issues.md"
        write(prompt_path, prompt)
        if args.execute:
            result = run(["codex", "exec", prompt], execute=True, capture=True, check=False)
            if result.returncode != 0:
                write(output_path.with_suffix(".error.txt"), result.stderr)
                log(f"Architect failed for {path.name}; see {output_path.with_suffix('.error.txt')}")
                continue
            write(output_path, result.stdout)
        else:
            # Deterministic fallback skeleton for dry-run/manual editing.
            title = read(path).splitlines()[0].replace("# Business Task:", "").strip() or path.stem
            labels = "bqa:arch-approved,bqa:ready-dev"
            if "Static" in title or "static" in body.lower():
                labels += ",bqa:static-site"
            if "Visualization" in title or "warcraft" in body.lower():
                labels += ",bqa:game-ui"
            if "Codex" in title or "pipeline" in body.lower():
                labels += ",bqa:codex-team"
            spec = ISSUE_TEMPLATE.format(
                context=f"Business task from `{path.name}`. This issue was routed through the technical architect stage.",
                goal=title,
                create_change="- To be refined by Architect/Codex output before execution.",
                do_not_touch="- Private repo data\n- Real session logs\n- Secrets",
                behavior=body,
                acceptance="- [ ] Architecture boundaries are respected.\n- [ ] Synthetic data only.\n- [ ] Manual verification steps pass.",
                verification="go test ./...",
            )
            write(output_path, f"---ISSUE---\nTITLE: {title}\nLABELS: {labels}\nBODY:\n{spec}\n---END---\n")
        state["processed_backlog"][path.name] = {"architected_at": now(), "output": str(output_path)}
        save_state(state)
        log(f"Architected {path.name} -> {output_path}")


def parse_issue_blocks(text: str) -> list[dict]:
    blocks = re.findall(r"---ISSUE---\s*(.*?)\s*---END---", text, re.S)
    issues = []
    for block in blocks:
        title_match = re.search(r"^TITLE:\s*(.+)$", block, re.M)
        labels_match = re.search(r"^LABELS:\s*(.+)$", block, re.M)
        body_match = re.search(r"^BODY:\s*(.*)$", block, re.M | re.S)
        if not title_match or not body_match:
            continue
        labels = []
        if labels_match:
            labels = [x.strip() for x in labels_match.group(1).split(",") if x.strip()]
        issues.append({"title": title_match.group(1).strip(), "labels": labels, "body": body_match.group(1).strip()})
    return issues


def cmd_ensure_labels(args: argparse.Namespace) -> None:
    require_tools(["gh"], args.execute)
    for label, desc in LABELS.items():
        run([
            "gh", "label", "create", label,
            "--repo", args.repo,
            "--description", desc,
            "--force",
        ], execute=args.execute, check=False)


def cmd_create_issues(args: argparse.Namespace) -> None:
    require_tools(["gh"], args.execute)
    state = load_state()
    issue_files = sorted(ISSUES_DIR.glob("*.issues.md"))
    if not issue_files:
        log("No generated issue specs found. Run `architect` first.")
        return
    for spec_file in issue_files:
        text = read(spec_file)
        for issue in parse_issue_blocks(text):
            key = f"{spec_file.name}:{issue['title']}"
            if key in state["created_issues"] and not args.force:
                log(f"Skip existing issue: {issue['title']}")
                continue
            body_file = GENERATED_DIR / "tmp" / f"{slugify(issue['title'])}.md"
            write(body_file, issue["body"])
            cmd = ["gh", "issue", "create", "--repo", args.repo, "--title", issue["title"], "--body-file", str(body_file)]
            for label in issue["labels"]:
                cmd += ["--label", label]
            result = run(cmd, execute=args.execute, capture=args.execute, check=False)
            url = result.stdout.strip() if args.execute else "DRY-RUN"
            state["created_issues"][key] = {"created_at": now(), "url": url, "labels": issue["labels"]}
            save_state(state)
            log(f"Issue ready: {issue['title']} -> {url}")


def issue_body(repo: str, number: int, execute: bool) -> str:
    result = run(["gh", "issue", "view", str(number), "--repo", repo, "--json", "title,body,labels", "--jq", "."], execute=execute, capture=True)
    if not execute:
        return ""
    return result.stdout


def dev_prompt(issue_json: str, repo: str) -> str:
    developer = load_role("developer")
    architect = load_role("architect")
    return f"""
{developer}

---

Architectural constraints:

{architect[:6000]}

---

Implement the following GitHub issue from `{repo}`.

Issue JSON:

{issue_json}

Workflow rules:
- Treat the issue as architecture-approved.
- Create a small, focused implementation.
- Follow AGENTS.md.
- Do not add private data, real session logs, or secrets.
- Add/update tests where reasonable.
- Run `go test ./...`.
- Do not merge anything.

At the end, summarize:
- files changed;
- tests run;
- any remaining risks.
""".strip()


def cmd_dev(args: argparse.Namespace) -> None:
    require_tools(["gh", "git", "codex"], args.execute)
    if not args.issue:
        raise SystemExit("dev requires --issue <number>")
    raw = issue_body(args.repo, args.issue, args.execute)
    if not args.execute:
        raw = json.dumps({"title": f"Issue {args.issue}", "body": "DRY-RUN", "labels": []}, indent=2)
    title = json.loads(raw).get("title", f"issue-{args.issue}") if raw else f"issue-{args.issue}"
    branch = args.branch or f"codex/issue-{args.issue}-{slugify(title, 32)}"
    prompt = dev_prompt(raw, args.repo)
    prompt_path = PROMPTS_DIR / f"dev_issue_{args.issue}.md"
    write(prompt_path, prompt)

    run(["git", "checkout", "-b", branch], execute=args.execute, check=False)
    run(["gh", "issue", "edit", str(args.issue), "--repo", args.repo, "--remove-label", "bqa:ready-dev", "--add-label", "bqa:in-dev"], execute=args.execute, check=False)
    result = run(["codex", "exec", prompt], execute=args.execute, capture=args.execute, check=False)
    if args.execute:
        write(GENERATED_DIR / "runs" / f"dev_issue_{args.issue}.out.txt", result.stdout + "\n" + result.stderr)
    run(["go", "test", "./..."], execute=args.execute, check=False)
    run(["git", "status", "--short"], execute=args.execute, check=False)
    if args.auto_commit:
        run(["git", "add", "."], execute=args.execute)
        run(["git", "commit", "-m", f"Implement issue #{args.issue}: {title}"], execute=args.execute, check=False)
        run(["git", "push", "-u", "origin", branch], execute=args.execute, check=False)
        pr_body = f"Implements #{args.issue}.\n\nGenerated by BQA Team Orchestrator using Developer role.\n\nChecklist:\n- [ ] go test ./... passes\n- [ ] QA review pending\n- [ ] Business acceptance pending\n"
        run(["gh", "pr", "create", "--repo", args.repo, "--title", f"Implement #{args.issue}: {title}", "--body", pr_body], execute=args.execute, check=False)
    run(["gh", "issue", "edit", str(args.issue), "--repo", args.repo, "--remove-label", "bqa:in-dev", "--add-label", "bqa:ready-qa"], execute=args.execute, check=False)


def qa_prompt(pr_number: int, repo: str) -> str:
    qa = load_role("qa")
    return f"""
{qa}

---

Review PR #{pr_number} in repository `{repo}` as BQA-OS QA / Test Engineer.

Tasks:
- inspect the PR diff;
- verify acceptance criteria from linked issue;
- run or recommend test commands;
- check architecture-sensitive QA risks;
- if implementation is incomplete, create a concise bug report body;
- do not approve weak work.

Return:
QA_STATUS: PASS or FAIL
BUG_TITLE: <only if FAIL>
BUG_BODY:
<only if FAIL>
""".strip()


def cmd_qa(args: argparse.Namespace) -> None:
    require_tools(["gh", "codex"], args.execute)
    if not args.pr:
        raise SystemExit("qa requires --pr <number>")
    prompt = qa_prompt(args.pr, args.repo)
    prompt_path = PROMPTS_DIR / f"qa_pr_{args.pr}.md"
    write(prompt_path, prompt)
    if args.execute:
        diff = run(["gh", "pr", "diff", str(args.pr), "--repo", args.repo], execute=True, capture=True, check=False).stdout
        prompt = prompt + "\n\nPR diff:\n\n" + diff[:30000]
        result = run(["codex", "exec", prompt], execute=True, capture=True, check=False)
        out = result.stdout + "\n" + result.stderr
    else:
        out = "QA_STATUS: DRY-RUN\n"
        print("DRY-RUN: would run Codex QA review")
    qa_out = GENERATED_DIR / "runs" / f"qa_pr_{args.pr}.out.txt"
    write(qa_out, out)
    log(f"QA output written: {qa_out}")
    if args.execute and "QA_STATUS: FAIL" in out:
        bug_title = re.search(r"BUG_TITLE:\s*(.+)", out)
        bug_body = re.search(r"BUG_BODY:\s*(.*)", out, re.S)
        title = bug_title.group(1).strip() if bug_title else f"QA bug found in PR #{args.pr}"
        body = bug_body.group(1).strip() if bug_body else out
        body_file = GENERATED_DIR / "tmp" / f"bug_pr_{args.pr}.md"
        write(body_file, body)
        run(["gh", "issue", "create", "--repo", args.repo, "--title", title, "--body-file", str(body_file), "--label", "bqa:bug", "--label", "bqa:qa-failed", "--label", "bqa:ready-dev"], execute=True, check=False)


def business_acceptance_prompt(pr_number: int, repo: str) -> str:
    business = load_role("business")
    return f"""
{business}

---

Perform final business acceptance for PR #{pr_number} in `{repo}`.

Evaluate:
- Does this deliver visible project value?
- Does it support the BQA-OS business direction?
- Is the UX/workflow understandable for first pilot users?
- Should this be accepted, revised, or split?

Return:
BUSINESS_STATUS: ACCEPT or REVISE
REASON:
<concise explanation>
""".strip()


def cmd_business_accept(args: argparse.Namespace) -> None:
    require_tools(["gh", "codex"], args.execute)
    if not args.pr:
        raise SystemExit("business-accept requires --pr <number>")
    prompt = business_acceptance_prompt(args.pr, args.repo)
    prompt_path = PROMPTS_DIR / f"business_accept_pr_{args.pr}.md"
    write(prompt_path, prompt)
    if args.execute:
        diff = run(["gh", "pr", "diff", str(args.pr), "--repo", args.repo], execute=True, capture=True, check=False).stdout
        result = run(["codex", "exec", prompt + "\n\nPR diff:\n\n" + diff[:30000]], execute=True, capture=True, check=False)
        out = result.stdout + "\n" + result.stderr
    else:
        out = "BUSINESS_STATUS: DRY-RUN\n"
        print("DRY-RUN: would run Codex business acceptance")
    output_path = GENERATED_DIR / "runs" / f"business_accept_pr_{args.pr}.out.txt"
    write(output_path, out)
    log(f"Business acceptance output written: {output_path}")


def list_ready_issues(repo: str, execute: bool, label: str = "bqa:ready-dev") -> list[int]:
    if not execute:
        return []
    result = run(["gh", "issue", "list", "--repo", repo, "--label", label, "--state", "open", "--json", "number", "--jq", ".[].[].number"], execute=True, capture=True, check=False)
    nums = []
    for line in result.stdout.splitlines():
        try:
            nums.append(int(line.strip()))
        except ValueError:
            pass
    return nums


def cmd_loop(args: argparse.Namespace) -> None:
    if not args.once and args.max_cycles <= 0:
        raise SystemExit("Refusing unbounded loop. Use --once or --max-cycles N.")
    cycles = 1 if args.once else args.max_cycles
    for i in range(cycles):
        log(f"Cycle {i + 1}/{cycles}")
        cmd_architect(args)
        cmd_create_issues(args)
        ready = list_ready_issues(args.repo, args.execute)
        if ready:
            log(f"Ready issues: {ready}")
        else:
            log("No ready-dev issues found or dry-run mode.")
        if args.sleep_seconds and i < cycles - 1:
            time.sleep(args.sleep_seconds)


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="BQA team role orchestrator")
    p.add_argument("--repo", default=os.environ.get("BQA_REPO", "mshegolev/bqa-os"), help="GitHub repo, e.g. mshegolev/bqa-os")
    p.add_argument("--execute", action="store_true", help="Actually run mutating commands and Codex. Default is dry-run.")
    sub = p.add_subparsers(dest="cmd", required=True)

    sub.add_parser("init")
    sub.add_parser("seed")
    sub.add_parser("ensure-labels")

    arch = sub.add_parser("architect")
    arch.add_argument("--force", action="store_true")

    ci = sub.add_parser("create-issues")
    ci.add_argument("--force", action="store_true")

    dev = sub.add_parser("dev")
    dev.add_argument("--issue", type=int, required=True)
    dev.add_argument("--branch")
    dev.add_argument("--auto-commit", action="store_true", help="Commit, push and open PR after Codex run")

    qa = sub.add_parser("qa")
    qa.add_argument("--pr", type=int, required=True)

    ba = sub.add_parser("business-accept")
    ba.add_argument("--pr", type=int, required=True)

    loop = sub.add_parser("loop")
    loop.add_argument("--once", action="store_true")
    loop.add_argument("--max-cycles", type=int, default=0)
    loop.add_argument("--sleep-seconds", type=int, default=0)
    loop.add_argument("--force", action="store_true")

    return p


def main() -> None:
    args = build_parser().parse_args()
    handlers = {
        "init": cmd_init,
        "seed": cmd_seed,
        "ensure-labels": cmd_ensure_labels,
        "architect": cmd_architect,
        "create-issues": cmd_create_issues,
        "dev": cmd_dev,
        "qa": cmd_qa,
        "business-accept": cmd_business_accept,
        "loop": cmd_loop,
    }
    handlers[args.cmd](args)


if __name__ == "__main__":
    main()
