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

function initUI(root = document) {
  initTargetForm(root);
  initTooltips(root);
}

document.addEventListener("DOMContentLoaded", () => initUI());
document.body.addEventListener("htmx:afterSettle", () => initUI());
