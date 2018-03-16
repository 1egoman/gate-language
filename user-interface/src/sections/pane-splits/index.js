import Split from 'split.js';
import './splits.css';

// Construct two splits:
// 1. A left split for the editor
// 2. A right split for the viewport
export default function buildSplits() {
  Split(['#editor-parent', '#viewport-parent'], {
    sizes: [25, 75],
    minSize: 400,
    onDrag() {
      const viewport = document.getElementById('viewport');
      const viewboxPrefix = viewport.getAttribute('viewBox').split(' ').slice(0, 2).join(' ');
      const viewboxZoom = window.parseInt(viewport.getAttribute('data-zoom'), 10);

      viewport.setAttribute(
        'viewBox',
        `${viewboxPrefix} ${viewboxZoom * viewport.clientWidth} ${viewboxZoom * viewport.clientHeight}`
      );
    },
  });
}
