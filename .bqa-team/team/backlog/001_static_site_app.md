# Business Task: Static BQA Web App MVP

Create a static HTML/JS application for BQA-OS where a user can upload a specially marked session archive and receive a generated output archive containing agents, workflows, specs, instructions, and recommendations.

Constraints:
- static site only: HTML/CSS/JavaScript;
- local-first: processing should happen in browser when feasible;
- no private data uploaded to external services by default;
- output should be downloadable as a zip archive;
- UX should explain the expected archive structure;
- should include sample synthetic input data only.

Expected outputs from the generated archive:
- agents/
- workflows/
- specs/
- knowledge/
- README_NEXT_STEPS.md
- recommendations.md
