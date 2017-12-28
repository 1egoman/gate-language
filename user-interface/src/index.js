import CodeMirror from 'codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/monokai.css';
import 'codemirror/mode/jsx/jsx';
import 'codemirror/addon/mode/simple';

import './index.css';

import renderViewport from './render';

import deepDiff from 'deep-diff';
import debounce from 'lodash.debounce';

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
  value: localStorage.source || `((1 and 0) or 0)`,
  mode: 'bitlang',
  theme: 'monokai',
});
editor.setSize('100%', '100%');

const compile = debounce(function compile(source) {
  return fetch('http://localhost:8080/v1', {
    method: 'POST',
    body: source,
    headers: {
      'Content-Type': 'text/plain',
      'Accept': 'application/json',
    },
  }).then(result => {
    if (result.ok) {
      return result.json();
    } else {
      throw new Error(`Compilation failed: ${result.statusCode}`);
    }
  }).then(newData => {
    // Was an error received while compiling?
    if (newData.Error) {
      throw new Error(newData.Error);
    }

    const oldData = data;

    // If there was previous data rendered in the viewport...
    if (oldData && oldData.Gates.length !== 0) {
      // ... Diff the old data and the new data. Apply all patches that don't change the position.
      deepDiff.observableDiff(oldData, newData, d => {
        if (!(d.path[d.path.length-1] === 'xPosition' || d.path[d.path.length-1] === 'yPosition')) {
          deepDiff.applyChange(oldData, newData, d);
        }
      });
      data = oldData;
    } else {
      // There's no old data to compare against, so the data is just the new data.
      data = newData;
    }

    data.Gates = data.Gates || []
    data.Wires = data.Wires || []

    // Randomize starting posititons of gates that don't have a position.
    newData.Gates.forEach(i => {
      if (i.xPosition && i.yPosition) {
        return
      }

      if (i.Inputs.length > 0 && i.Inputs[0].xPosition) {
        // Pick a position near an input
        i.xPosition = i.Inputs[0].xPosition + ((Math.random() * 100) - 50);
        i.yPosition = i.Inputs[0].yPosition + ((Math.random() * 100) - 50);
      } else if (i.Outputs.length > 0 && i.Outputs[0].xPosition) {
        // Pick a position near an output
        i.xPosition = i.Outputs[0].xPosition + ((Math.random() * 100) - 50);
        i.yPosition = i.Outputs[0].yPosition + ((Math.random() * 100) - 50);
      } else {
        i.xPosition = (Math.random() * 500);
        i.yPosition = (Math.random() * 500);
      }
    });

    localStorage.data = JSON.stringify(data);
    error = null;
  }).catch(err => {
    // Set a global error variable
    error = err.message;
  });
}, 1000);

let data = localStorage.data ? JSON.parse(localStorage.data) : {Gates: [], Wires: [], Outputs: []};
let error = null;
editor.on('change', () => {
  const value = editor.getValue();
  localStorage.source = value;
  compile(value);
});
compile(editor.getValue());

// Get a reference to the svg viewport
const viewport = document.getElementById('viewport');
const updateViewport = renderViewport(viewport);
function animate() {
  data.Gates = data.Gates || []
  data.Wires = data.Wires || []
  data.Outputs = data.Outputs || []

  updateViewport(data, error, {viewboxX, viewboxY});
  window.requestAnimationFrame(animate);
}
window.requestAnimationFrame(animate);


const zoomSlider = document.getElementById('zoom-slider');







let viewboxX = 0;
let viewboxY = 0;
let viewboxZoom = 1;

function zoomViewbox() {
  viewboxZoom = zoomSlider.value / 100; // 0 <= zoomSlider.value <= 100
  resizePanes(resizePosition);
}

zoomSlider.addEventListener('change', zoomViewbox);
zoomSlider.addEventListener('input', zoomViewbox);


// Adjust the position of the viewbox when the user drags around the svg canvas.
let moveOnSvg = false;
viewport.addEventListener('mousedown', event => {
  moveOnSvg = event.target.getAttribute('id') === 'viewport' || event.target.getAttribute('id') === 'block';

  // Deselect all gates if clicking on the viewport background or a block.
  if (moveOnSvg) {
    data.Gates.forEach(i => {
      i.active = false;
    });
  }
});
viewport.addEventListener('mousemove', event => {
  let selected = data.Gates.filter(i => i.active === true);
  if (event.buttons > 0 && moveOnSvg) {
    viewboxX -= viewboxZoom * event.movementX;
    viewboxY -= viewboxZoom * event.movementY;
    resizePanes(resizePosition);
  } else if (event.buttons > 0 && selected.length > 0) {
    selected.forEach(s => {
      s.xPosition = (s.xPosition || 0) + event.movementX;
      s.yPosition = (s.yPosition || 0) + event.movementY;
    })
  }

  // Save data to persistant localstorage.
  localStorage.data = JSON.stringify(data);
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
  viewport.setAttribute('viewBox', `${viewboxX} ${viewboxY} ${viewboxZoom * viewport.clientWidth} ${viewboxZoom * viewport.clientHeight}`);
}

resizeBar.addEventListener('mousemove', function(event) {
  if (event.buttons > 0) {
    resizePosition = event.screenX;
    resizePanes(resizePosition);
  }
});

// Initial svg size
resizePanes(resizePosition);
