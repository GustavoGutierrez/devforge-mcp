(function () {
  const search = document.getElementById('tool-search');
  const root = document.getElementById('tool-groups');
  const count = document.getElementById('tool-count');
  if (!search || !root || !window.DEVFORGE_TOOLS) return;

  function esc(str) {
    return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  function fmtJSON(raw) {
    try {
      const obj = JSON.parse(raw);
      const pretty = JSON.stringify(obj, null, 2);
      return esc(pretty).replace(
        /("(?:[^"\\]|\\.)*")\s*:/g,
        '<span class="json-key">$1</span>:'
      ).replace(
        /:\s*("(?:[^"\\]|\\.)*")/g,
        ': <span class="json-string">$1</span>'
      ).replace(
        /:\s*(true|false|null|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)/g,
        ': <span class="json-primitive">$1</span>'
      );
    } catch (_) {
      return esc(raw);
    }
  }

  function render(filter) {
    const q = (filter || '').toLowerCase().trim();
    const groups = {};
    let total = 0;
    window.DEVFORGE_TOOLS.forEach(function(t) {
      var hay = (t.name + ' ' + t.canonical + ' ' + t.category + ' ' + t.purpose + ' ' + t.inputs + ' ' + t.example).toLowerCase();
      if (q && hay.indexOf(q) === -1) return;
      groups[t.category] = groups[t.category] || [];
      groups[t.category].push(t);
      total += 1;
    });
    count.textContent = total + ' tools shown';

    var html = '';
    var cats = Object.entries(groups);
    for (var ci = 0; ci < cats.length; ci++) {
      var cat = cats[ci][0];
      var tools = cats[ci][1];
      html += '<section class="card"><h2>' + esc(cat) + '</h2><div class="tool-grid">';
      for (var ti = 0; ti < tools.length; ti++) {
        var t = tools[ti];
        html += '<article class="card tool-card">' +
          '<h3>' + esc(t.name) + '</h3>' +
          '<p class="muted">Canonical MCP: <code>' + esc(t.canonical) + '</code></p>' +
          '<p>' + esc(t.purpose) + '</p>' +
          '<p><span class="badge">Inputs</span> ' + esc(t.inputs) + '</p>' +
          '<pre class="code-json"><code>' + fmtJSON(t.example) + '</code></pre>' +
          '</article>';
      }
      html += '</div></section>';
    }
    root.innerHTML = html;
  }

  search.addEventListener('input', function(e) { render(e.target.value); });
  render('');
})();
