import Split from 'split.js';
import './splits.css';

// Construct two splits:
// 1. A left split for the editor
// 2. A right split for the viewport

Split(['#editor-parent', '#viewport-parent'], {
  sizes: [25, 75],
  minSize: 400,
});
