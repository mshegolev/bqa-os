const scenarios = {
  "B2B SaaS": {
    title: "Release readiness memory",
    copy:
      "Convert repeated release notes and regression checks into a reusable readiness workflow for every sprint.",
    inputs: "18 artifacts",
    output: "Release review",
  },
  Fintech: {
    title: "Risk-focused regression memory",
    copy:
      "Turn recurring payment, audit, and exception checks into a structured workflow for high-risk changes.",
    inputs: "24 artifacts",
    output: "Risk checklist",
  },
  "Data platform": {
    title: "Data quality review memory",
    copy:
      "Extract repeatable validation steps from pipeline notes, schema checks, and QA prompts.",
    inputs: "21 artifacts",
    output: "Validation workflow",
  },
};

const scenarioButtons = document.querySelectorAll(".scenario-button");
const scenarioLabel = document.querySelector("#scenario-label");
const scenarioTitle = document.querySelector("#scenario-title");
const scenarioCopy = document.querySelector("#scenario-copy");
const scenarioInputs = document.querySelector("#scenario-inputs");
const scenarioOutput = document.querySelector("#scenario-output");
const copyButton = document.querySelector("#copy-summary");
const copyStatus = document.querySelector("#copy-status");

scenarioButtons.forEach((button) => {
  button.addEventListener("click", () => {
    setScenario(button.textContent.trim());
  });
});

copyButton?.addEventListener("click", async () => {
  const summary =
    "BQA-OS 2-week QA Memory Pilot: provide 10-30 synthetic or sanitized QA artifacts; receive a reusable QA knowledge base plus 3-5 AI-assisted QA workflows.";

  try {
    await navigator.clipboard.writeText(summary);
    copyStatus.textContent = "Pilot summary copied.";
  } catch {
    copyStatus.textContent = summary;
  }
});

function setScenario(name) {
  const scenario = scenarios[name];
  if (!scenario) {
    return;
  }

  scenarioButtons.forEach((button) => {
    const active = button.textContent.trim() === name;
    button.classList.toggle("is-active", active);
    button.setAttribute("aria-pressed", String(active));
  });

  scenarioLabel.textContent = name;
  scenarioTitle.textContent = scenario.title;
  scenarioCopy.textContent = scenario.copy;
  scenarioInputs.textContent = scenario.inputs;
  scenarioOutput.textContent = scenario.output;
}
