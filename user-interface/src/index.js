import CodeMirror from 'codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/monokai.css';
import 'codemirror/mode/jsx/jsx';
import 'codemirror/addon/mode/simple';

import './index.css';

import renderViewport from './render';

import registerServiceWorker from './registerServiceWorker';
registerServiceWorker();

// Define editor parameters
CodeMirror.defineSimpleMode('bitlang', {
  // The start state contains the rules that are intially used
  start: [
    {regex: /(block)(\s+)([A-Za-z_][A-Za-z0-9_]*)/, token: ["keyword", null, "variable-2"]},
    {regex: /(let|return|block)\b/, token: "keyword"},
    {regex: /(?:1|0)/, token: "atom"},
    {regex: /\/\*/, token: "comment", next: "comment"},
    {regex: /\/\/[^\n]*/, token: "comment"},
    {regex: /(?:and|or|not)/, token: "property"},
    {regex: /[A-Za-z_][A-Za-z0-9_]*/, token: "variable-3"},
    {regex: /[{[(]/, indent: true},
    {regex: /[}\])]/, dedent: true},
  ],
  // The multi-line comment state.
  comment: [
    {regex: /.*?\*\//, token: "comment", next: "start"},
    {regex: /.*/, token: "comment"}
  ],
  meta: {
    dontIndentStates: ["comment"],
    lineComment: "//"
  },
});

// Create editor
const editor = CodeMirror(document.getElementById('editor-parent'), {
  value: `block half_add(a b) {
  let sum = (((not a) and b) or (b and (not a)))
  let carry = (a and b)
  return sum carry
}

half_add(1 0)`,
  mode: 'bitlang',
  theme: 'monokai',
});
editor.setSize('100%', '100%');

// Get a reference to the svg viewport
const viewport = document.getElementById('viewport');
const data = {"Gates":[{"Id":1,"Type":"SOURCE","Label":"","Inputs":[],"Outputs":[{"Id":1,"Desc":"","Start":null,"End":null}]},{"Id":2,"Type":"BLOCK_INPUT","Label":"Input 0 into block foo invocation 1","Inputs":[{"Id":1,"Desc":"","Start":null,"End":null}],"Outputs":[{"Id":2,"Desc":"","Start":null,"End":null}]},{"Id":3,"Type":"AND","Label":"","Inputs":[{"Id":2,"Desc":"","Start":null,"End":null},{"Id":2,"Desc":"","Start":null,"End":null}],"Outputs":[{"Id":3,"Desc":"","Start":null,"End":null}]},{"Id":4,"Type":"BLOCK_OUTPUT","Label":"Output 0 from block foo invocation 1","Inputs":[{"Id":2,"Desc":"","Start":null,"End":null}],"Outputs":[{"Id":4,"Desc":"","Start":null,"End":null}]},{"Id":5,"Type":"BLOCK_OUTPUT","Label":"Output 1 from block foo invocation 1","Inputs":[{"Id":3,"Desc":"","Start":null,"End":null}],"Outputs":[{"Id":5,"Desc":"","Start":null,"End":null}]}],"Wires":[{"Id":1,"Desc":"","Start":null,"End":null},{"Id":2,"Desc":"","Start":null,"End":null},{"Id":2,"Desc":"","Start":null,"End":null},{"Id":2,"Desc":"","Start":null,"End":null},{"Id":3,"Desc":"","Start":null,"End":null},{"Id":3,"Desc":"","Start":null,"End":null},{"Id":2,"Desc":"","Start":null,"End":null},{"Id":3,"Desc":"","Start":null,"End":null},{"Id":4,"Desc":"","Start":null,"End":null},{"Id":5,"Desc":"","Start":null,"End":null},{"Id":4,"Desc":"","Start":null,"End":null},{"Id":5,"Desc":"","Start":null,"End":null},{"Id":4,"Desc":"","Start":null,"End":null}],"Outputs":[{"Id":4,"Desc":"","Start":null,"End":null}]}

// Randomize starting posititons
data.Gates.forEach(i => {
  i.xPosition = Math.random() * 500;
  i.yPosition = Math.random() * 500;
});

const updateViewport = renderViewport(viewport);
function animate() {
  updateViewport(data);
  window.requestAnimationFrame(animate);
}
window.requestAnimationFrame(animate);










let viewboxX = 0;
let viewboxY = 0;

// Adjust the position of the viewbox when the user drags around the svg canvas.
let moveOnSvg = false;
viewport.addEventListener('mousedown', event => {
  moveOnSvg = event.target.getAttribute('id') === 'viewport';

  // Deselect all gates if clicking on the viewport background.
  if (moveOnSvg) {
    data.Gates.forEach(i => {
      i.active = false;
    });
  }
});
viewport.addEventListener('mousemove', event => {
  let selected = data.Gates.filter(i => i.active === true);
  if (event.buttons > 0 && moveOnSvg) {
    viewboxX -= event.movementX;
    viewboxY -= event.movementY;
    resizePanes(resizePosition);
  } else if (event.buttons > 0 && selected.length > 0) {
    selected.forEach(s => {
      s.xPosition = (s.xPosition || 0) + event.movementX;
      s.yPosition = (s.yPosition || 0) + event.movementY;
    })
  }
});


// Handle resizing of editor/viewport split.
const RESIZE_BAR_WIDTH = 50;
const resizeBar = document.getElementById('resize-bar');
let resizePosition = resizeBar.offsetLeft + (RESIZE_BAR_WIDTH / 2);

function resizePanes(resizePosition) {
  document.getElementById('editor-parent').style.width = `${resizePosition - (RESIZE_BAR_WIDTH / 2)}px`;
  viewport.setAttribute('width', `${document.body.clientWidth - resizePosition - (RESIZE_BAR_WIDTH / 2)}px`);
  viewport.setAttribute('height', `${document.body.clientHeight}px`);

  // If a viewbox has not been set, set it to `0 0 width height` (filling up the whole svg.)
  // Otherwise, adjust the viewbox width and height but keep the x and y coordinates the same.
  viewport.setAttribute('viewBox', `${viewboxX} ${viewboxY} ${viewport.clientWidth} ${viewport.clientHeight}`);
}

resizeBar.addEventListener('mousemove', function(event) {
  if (event.buttons > 0) {
    resizePosition = event.screenX;
    resizePanes(resizePosition);
  }
});

// Initial svg size
resizePanes(resizePosition);
