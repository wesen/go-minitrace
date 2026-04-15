(() => {
  const app = document.getElementById("app");
  const tooltip = document.getElementById("timeline-tooltip");
  if (!app) {
    return;
  }

  const payload = readPayload();
  if (!payload) {
    app.textContent = "Failed to load embedded timeline payload.";
    return;
  }

  const timeline = payload.timeline || {};
  const buckets = Array.isArray(timeline.buckets) ? timeline.buckets : [];
  const fileSeries = Array.isArray(timeline.file_series) ? timeline.file_series : [];
  const idleWindows = Array.isArray(timeline.idle_windows) ? timeline.idle_windows : [];
  const phaseMarkers = Array.isArray(timeline.phase_markers) ? timeline.phase_markers : [];
  const threadSpans = Array.isArray(timeline.thread_spans) ? timeline.thread_spans : [];

  const operationStyles = {
    read: { label: "Read", rgb: [96, 165, 250] },
    modify: { label: "Modify", rgb: [251, 146, 60] },
    new: { label: "New", rgb: [52, 211, 153] },
    execute: { label: "Execute", rgb: [167, 139, 250] },
  };

  let statusNode = null;

  renderHeader();
  renderHeatmap();
  renderFileBands();
  renderIdleWindows();
  renderPhaseRibbon();
  renderThreadBars();

  function readPayload() {
    const scriptTag = document.getElementById("minitrace-timeline-export-data");
    if (!scriptTag) {
      return null;
    }
    try {
      return JSON.parse(scriptTag.textContent || "{}");
    } catch (error) {
      console.error("Failed to parse timeline export payload", error);
      return null;
    }
  }

  function renderHeader() {
    const card = createCard();
    const top = div("topline");

    const titleWrap = document.createElement("div");
    const h1 = document.createElement("h1");
    const session = payload.session || {};
    h1.textContent = session.title || session.id || "Timeline export";
    titleWrap.appendChild(h1);

    const subtitle = document.createElement("p");
    subtitle.className = "help";
    const bucketMinutes = timeline.bucket_minutes || 30;
    subtitle.textContent = `Session ${session.id || "unknown"} • ${buckets.length} buckets @ ${bucketMinutes}m`;
    titleWrap.appendChild(subtitle);

    top.appendChild(titleWrap);
    card.appendChild(top);

    const meta = div("meta");
    meta.appendChild(metaItem("Started", session.started_at || "n/a"));
    meta.appendChild(metaItem("Ended", session.ended_at || "n/a"));
    meta.appendChild(metaItem("Turns", number(session.turn_count)));
    meta.appendChild(metaItem("Tool calls", number(session.tool_call_count)));
    meta.appendChild(metaItem("Annotations", number(session.annotation_count)));
    meta.appendChild(metaItem("Reader base URL", (payload.reader_links && payload.reader_links.base_url) || "(not set)"));
    card.appendChild(meta);

    const help = document.createElement("p");
    help.className = "help";
    help.textContent = "Hover any timeline element for details. Click to jump to the corresponding reader turn hash.";
    card.appendChild(help);

    statusNode = document.createElement("p");
    statusNode.className = "status";
    card.appendChild(statusNode);

    app.appendChild(card);
  }

  function renderHeatmap() {
    const card = createCard("Operation heatmap", "Bucketized operation intensity from timeline.buckets operation_counts.");
    if (!buckets.length) {
      card.appendChild(empty("No bucket data available."));
      app.appendChild(card);
      return;
    }

    const operations = ["read", "modify", "new", "execute"];
    const geom = createGeometry(buckets.length);
    const rowHeight = 24;
    const topPadding = 18;
    const bottomPadding = 22;
    const svgHeight = topPadding + operations.length * rowHeight + bottomPadding;
    const svg = createSVG(geom.totalWidth, svgHeight);

    const maxByOp = {};
    for (const key of operations) {
      maxByOp[key] = max(buckets.map((bucket) => count(bucket, key)));
    }

    operations.forEach((key, rowIdx) => {
      const y = topPadding + rowIdx * rowHeight;
      svg.appendChild(svgText(14, y + 9, `${operationStyles[key].label}`, "row-label"));

      buckets.forEach((bucket, bucketIdx) => {
        const x = geom.left + bucketIdx * geom.bucketWidth;
        const value = count(bucket, key);
        const maximum = maxByOp[key] || 1;
        const intensity = Math.min(1, value / maximum);
        const alpha = value <= 0 ? 0.08 : 0.22 + intensity * 0.78;
        const fill = rgba(operationStyles[key].rgb, alpha);
        const rect = svgRect(x, y, geom.bucketWidth - 1, rowHeight - 3, fill, "cell");
        bindInteractive(rect, [
          `${operationStyles[key].label} bucket ${bucketIdx}`,
          `${bucket.bucket_start || ""} → ${bucket.bucket_end || ""}`,
          `count: ${value}`,
          `jump: ${bucket.jump_hash || "(none)"}`,
        ].join("\n"), bucket.jump_hash || "");
        svg.appendChild(rect);
      });
    });

    svg.appendChild(svgText(geom.left, svgHeight - 7, "Buckets", "axis-label"));
    card.appendChild(svgWrap(svg));

    card.appendChild(legend([
      [rgba(operationStyles.read.rgb, 0.85), "Read"],
      [rgba(operationStyles.modify.rgb, 0.85), "Modify"],
      [rgba(operationStyles.new.rgb, 0.85), "New"],
      [rgba(operationStyles.execute.rgb, 0.85), "Execute"],
    ]));

    app.appendChild(card);
  }

  function renderFileBands() {
    const card = createCard("File activity bands", "Dense file-series bands from timeline.file_series (top-ranked files). Click active buckets to jump.");
    if (!buckets.length || !fileSeries.length) {
      card.appendChild(empty("No file activity series available."));
      app.appendChild(card);
      return;
    }

    const displayedRows = fileSeries.slice(0, 16);
    const geom = createGeometry(buckets.length);
    const rowHeight = 22;
    const topPadding = 14;
    const svgHeight = topPadding + displayedRows.length * rowHeight + 12;
    const svg = createSVG(geom.totalWidth, svgHeight);

    displayedRows.forEach((series, rowIdx) => {
      const y = topPadding + rowIdx * rowHeight;
      const label = `${basename(series.file_path || "(unknown)")} (${number(series.total_operations)})`;
      svg.appendChild(svgText(10, y + 8, shorten(label, 34), "row-label"));

      for (let bucketIdx = 0; bucketIdx < buckets.length; bucketIdx++) {
        const value = (series.series && series.series[bucketIdx]) || 0;
        if (value <= 0) {
          continue;
        }
        const dominant = ((series.dominant_operation_by_bucket && series.dominant_operation_by_bucket[bucketIdx]) || "").toLowerCase();
        const style = operationStyles[dominant] || { rgb: [226, 232, 240], label: dominant || "activity" };
        const alpha = Math.min(0.95, 0.25 + value / 8);
        const x = geom.left + bucketIdx * geom.bucketWidth;
        const rect = svgRect(x, y, geom.bucketWidth - 1, rowHeight - 4, rgba(style.rgb, alpha), "cell");
        const bucket = buckets[bucketIdx] || {};
        bindInteractive(rect, [
          `File: ${series.file_path}`,
          `Bucket ${bucketIdx}`,
          `Operations: ${value}`,
          `Dominant: ${dominant || "unknown"}`,
          `jump: ${bucket.jump_hash || "(none)"}`,
        ].join("\n"), bucket.jump_hash || "");
        svg.appendChild(rect);
      }
    });

    card.appendChild(svgWrap(svg));
    card.appendChild(legend([
      [rgba(operationStyles.read.rgb, 0.85), "Read-heavy bucket"],
      [rgba(operationStyles.modify.rgb, 0.85), "Modify-heavy bucket"],
      [rgba(operationStyles.new.rgb, 0.85), "New-file bucket"],
      [rgba(operationStyles.execute.rgb, 0.85), "Execute-heavy bucket"],
    ]));
    app.appendChild(card);
  }

  function renderIdleWindows() {
    const card = createCard("Idle window shading", "Contiguous idle spans from timeline.idle_windows overlayed on bucket axis.");
    if (!buckets.length) {
      card.appendChild(empty("No bucket axis available for idle windows."));
      app.appendChild(card);
      return;
    }

    const geom = createGeometry(buckets.length);
    const svg = createSVG(geom.totalWidth, 78);
    const y = 26;

    buckets.forEach((_, bucketIdx) => {
      const x = geom.left + bucketIdx * geom.bucketWidth;
      svg.appendChild(svgRect(x, y, geom.bucketWidth - 1, 22, "rgba(100, 116, 139, 0.14)", ""));
    });

    idleWindows.forEach((window) => {
      const x = geom.left + window.start_bucket_idx * geom.bucketWidth;
      const width = (window.end_bucket_idx - window.start_bucket_idx + 1) * geom.bucketWidth - 1;
      const rect = svgRect(x, y - 1, Math.max(2, width), 24, "var(--idle)", "cell");
      const jumpHash = bucketJump(window.start_bucket_idx);
      bindInteractive(rect, [
        `Idle window ${window.start_bucket_idx}-${window.end_bucket_idx}`,
        `${window.idle_start || ""} → ${window.idle_end || ""}`,
        `Buckets: ${number(window.bucket_count)}`,
        `Tool calls: ${number(window.tool_call_count_in_window)}`,
        `jump: ${jumpHash || "(none)"}`,
      ].join("\n"), jumpHash);
      svg.appendChild(rect);

      if (width > 46) {
        svg.appendChild(svgText(x + 4, y + 14, `${window.bucket_count} idle`, "row-subtext"));
      }
    });

    svg.appendChild(svgText(geom.left, 66, "Bucket index increases left → right", "axis-label"));
    card.appendChild(svgWrap(svg));
    card.appendChild(legend([["var(--idle)", "Idle window"]]));
    app.appendChild(card);
  }

  function renderPhaseRibbon() {
    const card = createCard("Phase ribbon", "Merged phase markers from manual/imported annotations and SQL-derived phase signals.");
    if (!buckets.length || !phaseMarkers.length) {
      card.appendChild(empty("No phase markers available."));
      app.appendChild(card);
      return;
    }

    const geom = createGeometry(buckets.length);
    const svg = createSVG(geom.totalWidth, 88);
    const y = 28;

    phaseMarkers.forEach((phase) => {
      const start = safeInt(phase.start_bucket_idx, 0);
      const end = safeInt(phase.end_bucket_idx, start);
      const x = geom.left + start * geom.bucketWidth;
      const width = (end - start + 1) * geom.bucketWidth - 1;
      const fill = phase.source === "manual" ? "rgba(245, 158, 11, 0.75)" : "rgba(56, 189, 248, 0.66)";
      const rect = svgRect(x, y, Math.max(2, width), 24, fill, "cell");
      const jumpHash = phase.jump_hash || bucketJump(start);
      bindInteractive(rect, [
        `${phase.label || phase.phase_id}`,
        `Buckets: ${start}-${end}`,
        `Source: ${phase.source || "unknown"}`,
        phase.phase_signal ? `Signal: ${phase.phase_signal}` : "",
        `jump: ${jumpHash || "(none)"}`,
      ].filter(Boolean).join("\n"), jumpHash);
      svg.appendChild(rect);

      if (width > 44) {
        svg.appendChild(svgText(x + 4, y + 15, shorten(phase.label || phase.phase_id || "phase", Math.max(6, Math.floor(width / 8))), "row-subtext"));
      }
    });

    card.appendChild(svgWrap(svg));
    card.appendChild(legend([
      ["rgba(245, 158, 11, 0.8)", "Manual/imported phase"],
      ["rgba(56, 189, 248, 0.7)", "SQL-derived phase"],
    ]));
    app.appendChild(card);
  }

  function renderThreadBars() {
    const card = createCard("Thread bars", "Segmented thread spans from timeline.thread_spans (manual/imported + SQL-derived). Click segments to jump.");
    if (!buckets.length || !threadSpans.length) {
      card.appendChild(empty("No thread spans available."));
      app.appendChild(card);
      return;
    }

    const rows = threadSpans.slice(0, 18);
    const geom = createGeometry(buckets.length);
    const rowHeight = 24;
    const topPadding = 12;
    const svgHeight = topPadding + rows.length * rowHeight + 10;
    const svg = createSVG(geom.totalWidth, svgHeight);

    rows.forEach((span, rowIdx) => {
      const y = topPadding + rowIdx * rowHeight;
      const label = `${span.label || span.thread_id || "thread"}${span.file_path ? ` (${basename(span.file_path)})` : ""}`;
      svg.appendChild(svgText(10, y + 8, shorten(label, 34), "row-label"));

      const segments = Array.isArray(span.segments) ? span.segments : [];
      segments.forEach((segment) => {
        const start = safeInt(segment.start_bucket_idx, 0);
        const end = safeInt(segment.end_bucket_idx, start);
        const x = geom.left + start * geom.bucketWidth;
        const width = (end - start + 1) * geom.bucketWidth - 1;
        const fill = span.source === "manual" ? "rgba(245, 158, 11, 0.76)" : "rgba(125, 211, 252, 0.62)";
        const rect = svgRect(x, y, Math.max(2, width), rowHeight - 6, fill, "cell");
        const jumpHash = span.jump_hash || bucketJump(start);
        bindInteractive(rect, [
          `${span.label || span.thread_id}`,
          `Segment: ${start}-${end}`,
          `Source: ${span.source || "unknown"}`,
          span.file_path ? `File: ${span.file_path}` : "",
          span.total_operations ? `Operations: ${span.total_operations}` : "",
          `jump: ${jumpHash || "(none)"}`,
        ].filter(Boolean).join("\n"), jumpHash);
        svg.appendChild(rect);
      });
    });

    card.appendChild(svgWrap(svg));
    card.appendChild(legend([
      ["rgba(245, 158, 11, 0.8)", "Manual/imported thread"],
      ["rgba(125, 211, 252, 0.7)", "SQL-derived thread"],
    ]));
    app.appendChild(card);
  }

  function createGeometry(bucketCount) {
    const bucketWidth = bucketCount <= 40 ? 22 : bucketCount <= 120 ? 14 : 10;
    const left = 240;
    const right = 14;
    return {
      bucketWidth,
      left,
      totalWidth: left + bucketWidth * bucketCount + right,
    };
  }

  function bucketJump(bucketIdx) {
    const bucket = buckets[bucketIdx];
    return (bucket && bucket.jump_hash) || "";
  }

  function bindInteractive(node, text, jumpHash) {
    if (!node) {
      return;
    }
    node.setAttribute("tabindex", "0");
    node.addEventListener("mousemove", (event) => showTooltip(event.clientX, event.clientY, text));
    node.addEventListener("mouseleave", hideTooltip);
    node.addEventListener("focus", () => {
      const rect = node.getBoundingClientRect();
      showTooltip(rect.left + rect.width / 2, rect.top, text);
    });
    node.addEventListener("blur", hideTooltip);
    if (jumpHash) {
      node.addEventListener("click", () => jumpToReader(jumpHash));
      node.addEventListener("keydown", (event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          jumpToReader(jumpHash);
        }
      });
    }
  }

  function jumpToReader(hash) {
    if (!hash) {
      return;
    }
    const baseURL = ((payload.reader_links && payload.reader_links.base_url) || "").trim();
    if (!baseURL) {
      window.location.hash = hash;
      setStatus(`Set hash to ${hash}. Use --reader-base-url on export to open a specific reader file.`);
      return;
    }
    const href = baseURL.replace(/#.*/, "") + hash;
    setStatus(`Opening ${href}`);
    window.location.href = href;
  }

  function setStatus(message) {
    if (statusNode) {
      statusNode.textContent = message;
    }
  }

  function showTooltip(clientX, clientY, text) {
    if (!tooltip || !text) {
      return;
    }
    tooltip.textContent = text;
    tooltip.hidden = false;
    const padding = 14;
    const width = tooltip.offsetWidth || 320;
    const height = tooltip.offsetHeight || 80;
    const left = Math.min(window.innerWidth - width - padding, Math.max(padding, clientX + 14));
    const top = Math.min(window.innerHeight - height - padding, Math.max(padding, clientY + 14));
    tooltip.style.left = `${left}px`;
    tooltip.style.top = `${top}px`;
  }

  function hideTooltip() {
    if (tooltip) {
      tooltip.hidden = true;
    }
  }

  function createCard(title, subtitle) {
    const card = document.createElement("section");
    card.className = "card";
    if (title) {
      const head = document.createElement("div");
      head.className = "section-head";
      const h2 = document.createElement("h2");
      h2.textContent = title;
      head.appendChild(h2);
      if (subtitle) {
        const p = document.createElement("p");
        p.textContent = subtitle;
        head.appendChild(p);
      }
      card.appendChild(head);
    }
    return card;
  }

  function svgWrap(svg) {
    const wrap = div("svg-wrap");
    wrap.appendChild(svg);
    return wrap;
  }

  function legend(items) {
    const container = div("legend");
    items.forEach(([color, label]) => {
      const item = div("legend-item");
      const swatch = document.createElement("span");
      swatch.className = "swatch";
      swatch.style.background = color;
      item.appendChild(swatch);
      const text = document.createElement("span");
      text.textContent = label;
      item.appendChild(text);
      container.appendChild(item);
    });
    return container;
  }

  function empty(text) {
    const node = document.createElement("p");
    node.className = "empty";
    node.textContent = text;
    return node;
  }

  function createSVG(width, height) {
    const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    svg.setAttribute("width", String(width));
    svg.setAttribute("height", String(height));
    svg.setAttribute("viewBox", `0 0 ${width} ${height}`);
    svg.setAttribute("role", "img");
    return svg;
  }

  function svgRect(x, y, width, height, fill, className) {
    const node = document.createElementNS("http://www.w3.org/2000/svg", "rect");
    node.setAttribute("x", String(x));
    node.setAttribute("y", String(y));
    node.setAttribute("width", String(width));
    node.setAttribute("height", String(height));
    node.setAttribute("fill", fill);
    if (className) {
      node.setAttribute("class", className);
    }
    node.setAttribute("rx", "2");
    return node;
  }

  function svgText(x, y, text, className) {
    const node = document.createElementNS("http://www.w3.org/2000/svg", "text");
    node.setAttribute("x", String(x));
    node.setAttribute("y", String(y));
    if (className) {
      node.setAttribute("class", className);
    }
    node.textContent = text;
    return node;
  }

  function metaItem(label, value) {
    const item = div("meta-item");
    const span = document.createElement("span");
    span.textContent = label;
    item.appendChild(span);
    const strong = document.createElement("strong");
    strong.textContent = value;
    item.appendChild(strong);
    return item;
  }

  function div(className) {
    const node = document.createElement("div");
    node.className = className;
    return node;
  }

  function count(bucket, key) {
    if (!bucket || !bucket.operation_counts) {
      return 0;
    }
    return safeInt(bucket.operation_counts[key], 0);
  }

  function safeInt(value, fallback) {
    const parsed = Number(value);
    if (!Number.isFinite(parsed)) {
      return fallback;
    }
    return parsed;
  }

  function rgba(rgb, alpha) {
    return `rgba(${rgb[0]}, ${rgb[1]}, ${rgb[2]}, ${alpha.toFixed(3)})`;
  }

  function shorten(value, maxLen) {
    if (!value) {
      return "";
    }
    if (value.length <= maxLen) {
      return value;
    }
    return `${value.slice(0, Math.max(1, maxLen - 1))}…`;
  }

  function basename(path) {
    if (!path) {
      return "";
    }
    const normalized = String(path).replace(/\\/g, "/");
    const pieces = normalized.split("/");
    return pieces[pieces.length - 1] || normalized;
  }

  function max(values) {
    if (!values.length) {
      return 0;
    }
    return values.reduce((acc, value) => (value > acc ? value : acc), values[0]);
  }

  function number(value) {
    const parsed = Number(value);
    if (!Number.isFinite(parsed)) {
      return "0";
    }
    return parsed.toLocaleString("en-US");
  }
})();
