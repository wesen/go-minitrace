(function () {
  const root = document.getElementById('root');
  const payloadEl = document.getElementById('minitrace-export-data');
  if (!root || !payloadEl || !payloadEl.textContent) {
    return;
  }
  const data = JSON.parse(payloadEl.textContent);
  const annotationById = new Map((data.annotations || []).map(a => [a.id, a]));

  function annotationsForTurn(turnIdx) {
    const ids = (data.indices && data.indices.turn_to_annotations && data.indices.turn_to_annotations[String(turnIdx)]) || [];
    return ids.map(id => annotationById.get(id)).filter(Boolean);
  }

  function annotationsForTool(toolId) {
    const ids = (data.indices && data.indices.tool_call_to_annotations && data.indices.tool_call_to_annotations[toolId]) || [];
    return ids.map(id => annotationById.get(id)).filter(Boolean);
  }

  function renderAnnotation(a) {
    const tags = ((a.content && a.content.tags) || []).map(t => `<span class="badge">${escapeHtml(t)}</span>`).join('');
    return `<div class="annotation"><div class="ann-head">${escapeHtml(a.content.category)} · ${escapeHtml(a.content.title || '')}</div><div>${escapeHtml(a.content.detail || '')}</div>${tags ? `<div class="badges" style="margin-top:6px">${tags}</div>` : ''}</div>`;
  }

  function renderTool(tc) {
    const ann = annotationsForTool(tc.id).map(renderAnnotation).join('');
    const filePath = tc.input && tc.input.file_path ? ` · ${escapeHtml(tc.input.file_path)}` : '';
    const status = tc.output && tc.output.success ? 'success' : 'error';
    const body = tc.output && tc.output.error
      ? `<pre class="tool-error">${escapeHtml(tc.output.error)}</pre>`
      : (tc.output && tc.output.result ? `<pre>${escapeHtml(tc.output.result)}</pre>` : '');
    return `<details class="tool"><summary><strong>${escapeHtml(tc.tool_name)}</strong> · ${escapeHtml(tc.operation_type)}${filePath} · <span class="${status}">${status}</span></summary>${body}${ann ? `<div class="annotation-list" style="padding:10px">${ann}</div>` : ''}</details>`;
  }

  function renderTurn(turn) {
    const toolCalls = (turn.tool_calls_in_turn || []).map(renderTool).join('');
    const anns = annotationsForTurn(turn.idx).map(renderAnnotation).join('');
    return `<section class="turn ${escapeHtml(turn.role)}" id="turn-${turn.idx}"><div class="turn-head"><strong>${escapeHtml(turn.role)}</strong><span class="turn-id">#${turn.idx}</span>${turn.timestamp ? `<span class="ts">${escapeHtml(turn.timestamp)}</span>` : ''}</div><div class="turn-content">${escapeHtml(turn.content || '')}</div>${anns ? `<div class="annotation-list">${anns}</div>` : ''}${toolCalls ? `<div class="tool-list">${toolCalls}</div>` : ''}</section>`;
  }

  function renderBlock(block) {
    const turns = (block.turns || []).map(renderTurn).join('');
    return `<details class="block" open><summary><div class="block-title">Prompt #${block.block_num}</div><div class="block-stats">user turn #${block.user_turn_idx} · ${block.agent_turns} agent turns · ${block.tool_calls} tool calls</div><div style="margin-top:8px; color:#d9e2f1; white-space:pre-wrap">${escapeHtml(block.user_content || '')}</div></summary><div class="block-body">${turns}</div></details>`;
  }

  function render(filteredBlocks) {
    const session = data.session || {};
    root.innerHTML = `<div class="header"><h1>${escapeHtml(session.title || session.id || 'Transcript export')}</h1><div class="meta"><span>${escapeHtml(session.environment?.agent_framework || '')}</span><span>${escapeHtml(session.environment?.model || '')}</span><span>${escapeHtml(session.operational_context?.working_directory || '')}</span><span>${escapeHtml(session.id || '')}</span></div></div><div class="controls"><input id="search-box" type="search" placeholder="Search turns, tools, file paths, and outputs" value="${escapeHtml(currentQuery)}" /></div><div id="blocks">${filteredBlocks.map(renderBlock).join('')}</div>`;
    const searchBox = document.getElementById('search-box');
    if (searchBox) searchBox.addEventListener('input', (e) => { currentQuery = e.target.value || ''; applySearch(); });
  }

  function turnMatches(turn, q) {
    const query = q.toLowerCase();
    if ((turn.content || '').toLowerCase().includes(query)) return true;
    if ((turn.thinking || '').toLowerCase().includes(query)) return true;
    return (turn.tool_calls_in_turn || []).some(tc => {
      return (tc.tool_name || '').toLowerCase().includes(query)
        || ((tc.input && tc.input.file_path) || '').toLowerCase().includes(query)
        || ((tc.input && tc.input.command) || '').toLowerCase().includes(query)
        || ((tc.output && tc.output.result) || '').toLowerCase().includes(query)
        || ((tc.output && tc.output.error) || '').toLowerCase().includes(query);
    });
  }

  let currentQuery = '';
  function applySearch() {
    const q = currentQuery.trim();
    if (!q) {
      render(data.session.blocks || []);
      return;
    }
    const filtered = (data.session.blocks || []).map(block => ({
      ...block,
      turns: (block.turns || []).filter(turn => turnMatches(turn, q))
    })).filter(block => block.turns.length > 0 || ((block.user_content || '').toLowerCase().includes(q.toLowerCase())));
    render(filtered);
  }

  function escapeHtml(value) {
    return String(value)
      .replaceAll('&', '&amp;')
      .replaceAll('<', '&lt;')
      .replaceAll('>', '&gt;')
      .replaceAll('"', '&quot;')
      .replaceAll("'", '&#39;');
  }

  applySearch();
})();
