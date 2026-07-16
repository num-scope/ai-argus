import htmx from "htmx.org";
import Tooltip from "flowbite/lib/esm/components/tooltip/index.js";

window.htmx = htmx;

function initTargetForm(root = document) {
  const protocol = root.querySelector("[data-protocol]");
  const customFields = root.querySelector("[data-custom-fields]");
  const targetAdvanced = root.querySelector("[data-target-advanced]");
  if (!protocol || !customFields || protocol.dataset.bound) return;

  const syncProtocol = () => {
    const custom = protocol.value === "custom";
    customFields.hidden = !custom;
    if (custom && targetAdvanced) targetAdvanced.open = true;
  };
  protocol.dataset.bound = "true";
  protocol.addEventListener("change", syncProtocol);
  syncProtocol();
}

function initTooltips(root = document) {
  root.querySelectorAll(".field-help[data-help]").forEach((button, index) => {
    if (button.dataset.tooltipBound) return;

    const tooltip = document.createElement("div");
    const arrow = document.createElement("div");
    const id = `field-tooltip-${Date.now()}-${index}`;
    tooltip.id = id;
    tooltip.setAttribute("role", "tooltip");
    tooltip.setAttribute("aria-hidden", "true");
    tooltip.className = "argus-tooltip invisible absolute z-[2147483647] inline-block max-w-[300px] rounded-lg bg-neutral-900 px-3 py-2 text-left text-xs font-normal leading-relaxed text-white opacity-0 shadow-lg transition-opacity duration-150";
    tooltip.textContent = button.dataset.help;
    arrow.className = "tooltip-arrow";
    arrow.setAttribute("data-popper-arrow", "");
    tooltip.appendChild(arrow);
    document.body.appendChild(tooltip);

    const instance = new Tooltip(tooltip, button, {
      placement: "top",
      triggerType: "none",
      onShow: () => {
        button.setAttribute("aria-describedby", id);
        tooltip.setAttribute("aria-hidden", "false");
      },
      onHide: () => {
        button.removeAttribute("aria-describedby");
        tooltip.setAttribute("aria-hidden", "true");
      },
    });
    button.dataset.tooltipBound = "true";
    button.addEventListener("mouseenter", () => instance.show());
    button.addEventListener("mouseleave", () => instance.hide());
    button.addEventListener("click", (event) => {
      event.preventDefault();
      event.stopPropagation();
    });
  });
}

function initScenarioForm(root = document) {
  const form = root.querySelector("[data-scenario-form]");
  if (!form || form.dataset.bound) return;
  form.dataset.bound = "true";

  const randomToggle = form.querySelector("[data-random-prompt]");
  const randomFields = form.querySelector("[data-random-fields]");
  const promptsInput = form.querySelector("[data-prompts-input]");
  const promptFile = form.querySelector("[data-prompt-file]");

  const syncRandom = () => {
    const on = Boolean(randomToggle?.checked);
    if (randomFields) randomFields.hidden = !on;
    if (promptsInput) {
      promptsInput.required = !on;
      promptsInput.closest("[data-prompt-field]")?.classList.toggle("is-optional", on);
    }
  };

  randomToggle?.addEventListener("change", syncRandom);
  syncRandom();

  promptFile?.addEventListener("change", async () => {
    const file = promptFile.files?.[0];
    if (!file || !promptsInput) return;
    const text = await file.text();
    const lines = text
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter(Boolean);
    promptsInput.value = lines.join("\n");
    promptFile.value = "";
  });
}

function initUI(root = document) {
  initTargetForm(root);
  initScenarioForm(root);
  initTooltips(root);
}

function initPrintPDF() {
  // Ensure print layout uses full document flow; charts keep current canvas state.
  window.addEventListener("beforeprint", () => {
    document.documentElement.classList.add("is-printing");
  });
  window.addEventListener("afterprint", () => {
    document.documentElement.classList.remove("is-printing");
  });
}

document.addEventListener("DOMContentLoaded", () => {
  initUI();
  initPrintPDF();
});
document.body.addEventListener("htmx:afterSettle", () => initUI());
