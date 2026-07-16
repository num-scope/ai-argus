import * as echarts from "echarts/core";
import { BarChart } from "echarts/charts";
import { GridComponent, TooltipComponent } from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";

echarts.use([BarChart, GridComponent, TooltipComponent, CanvasRenderer]);

const chartInstances = new WeakMap();

function initRequestCharts(root = document) {
  root.querySelectorAll("[data-request-chart]").forEach((chart) => {
    if (chartInstances.has(chart)) return;

    const points = [...chart.querySelectorAll("[data-request-point]")]
      .map((point) => ({
        index: Number(point.dataset.index),
        elapsed: Number(point.dataset.elapsed),
        ok: point.dataset.ok === "true",
      }))
      .reverse();
    chart.replaceChildren();

    const instance = echarts.init(chart, null, { renderer: "canvas" });
    instance.setOption({
      animationDuration: 240,
      grid: { top: 14, right: 12, bottom: 28, left: 58 },
      tooltip: {
        trigger: "item",
        borderWidth: 0,
        backgroundColor: "#151515",
        textStyle: { color: "#f8f8f8", fontSize: 12 },
        formatter: ({ data }) => `请求 #${data.index}<br>${data.value.toFixed(1)} ms${data.ok ? "" : " · 失败"}`,
      },
      xAxis: {
        type: "category",
        data: points.map((point) => point.index),
        axisLine: { lineStyle: { color: "#d8d8d8" } },
        axisTick: { show: false },
        axisLabel: { color: "#888888", fontSize: 9, hideOverlap: true },
      },
      yAxis: {
        type: "value",
        min: 0,
        splitNumber: 3,
        axisLabel: { color: "#888888", fontSize: 9, formatter: "{value} ms" },
        splitLine: { lineStyle: { color: "#ececec" } },
      },
      series: [{
        type: "bar",
        barMaxWidth: 18,
        data: points.map((point) => ({
          value: point.elapsed,
          index: point.index,
          ok: point.ok,
          itemStyle: { color: point.ok ? "#171717" : "#e00", borderRadius: [2, 2, 0, 0] },
        })),
      }],
    });

    const observer = new ResizeObserver(() => instance.resize());
    observer.observe(chart);
    chartInstances.set(chart, { instance, observer });
  });
}

function disposeCharts(root) {
  const charts = root.matches?.("[data-request-chart]")
    ? [root]
    : [...root.querySelectorAll?.("[data-request-chart]") || []];
  charts.forEach((chart) => {
    const managed = chartInstances.get(chart);
    if (!managed) return;
    managed.observer.disconnect();
    managed.instance.dispose();
    chartInstances.delete(chart);
  });
}

document.addEventListener("DOMContentLoaded", () => {
  initRequestCharts();
  const contentObserver = new MutationObserver((mutations) => {
    let hasNewChart = false;
    mutations.forEach((mutation) => {
      mutation.removedNodes.forEach((node) => {
        if (node.nodeType === Node.ELEMENT_NODE) disposeCharts(node);
      });
      mutation.addedNodes.forEach((node) => {
        if (node.nodeType !== Node.ELEMENT_NODE) return;
        if (node.matches("[data-request-chart]") || node.querySelector("[data-request-chart]")) {
          hasNewChart = true;
        }
      });
    });
    if (hasNewChart) initRequestCharts();
  });
  contentObserver.observe(document.body, { childList: true, subtree: true });
});
