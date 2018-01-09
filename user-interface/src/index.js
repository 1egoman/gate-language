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
  lineNumbers: true,
  // value: `let a = toggle()\nled(a)`,
  value: `let a = toggle()\nlet b = toggle()\nled(a and b)`,
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
      // ... Diff the old data and the new data.
      deepDiff.observableDiff(oldData, newData, d => {
        // Don't apply patches that change the gate's position.
        if (d.path[d.path.length-1] === 'xPosition' || d.path[d.path.length-1] === 'yPosition') {
          return
        // Don't apply patches that relate to the powered state of a wire.
        } else if (d.path[d.path.length-1] === 'powered' || d.path[d.path.length-1] === 'powered') {
          return
        } else {
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

    error = null;
    renderFrame(data.Gates);
  }).catch(err => {
    // Set a global error variable
    error = err.message;
  });
}, 1000);

let data = {Gates: [], Wires: [], Outputs: []};
let error = null;
editor.on('change', () => {
  const value = editor.getValue();
  compile(value);
});
compile(editor.getValue());

// Get a reference to the svg viewport
const viewport = document.getElementById('viewport');
const updateViewport = renderViewport(viewport);
function renderFrame(changedGateIds) {
  data.Gates = data.Gates || []
  data.Wires = data.Wires || []
  data.Outputs = data.Outputs || []

  function setWire(id, powered) {
    const wire = data.Wires.find(i => i.Id === id);
    if (wire) {
      wire.powered = powered;
    }
  }

  function getWire(id) {
    const wire = data.Wires.find(i => i.Id === id);
    if (wire) {
      return wire.powered;
    }
  }

  // Update the powered state of the wires.
  while (true) {
    const initialState = JSON.stringify(data.Wires);

    data.Gates.forEach(gate => {
      switch (gate.Type) {
        case 'AND':
          setWire(gate.Outputs[0].Id, getWire(gate.Inputs[0].Id) && getWire(gate.Inputs[1].Id));
          break;
        case 'OR':
          setWire(gate.Outputs[0].Id, getWire(gate.Inputs[0].Id) || getWire(gate.Inputs[1].Id));
          break;
        case 'NOT':
          setWire(gate.Outputs[0].Id, !getWire(gate.Inputs[0].Id));
          break;

        case 'BLOCK_INPUT':
        case 'BLOCK_OUTPUT':
          setWire(gate.Outputs[0].Id, getWire(gate.Inputs[0].Id));
          break;

        case 'SOURCE':
          setWire(gate.Outputs[0].Id, true);
          break;
        case 'GROUND':
          setWire(gate.Outputs[0].Id, false);
          break;

        case 'BUILTIN_FUNCTION':
          if (['momentary', 'toggle'].indexOf(gate.Label) >= 0) {
            for (let i = 0; i < gate.Outputs.length; i++) {
              setWire(gate.Outputs[i].Id, gate.state === 'on');
            }
          } else if (['led'].indexOf(gate.Label) >= 0) {
            if (getWire(gate.Inputs[0].Id) === true) {
              gate.state = 'on';
            } else {
              gate.state = 'off';
            }
          }
          break;

        default:
          break;
      }
    });

    if (initialState === JSON.stringify(data.Wires)) {
      break;
    }
  }

  // Rerender the viewport.
  updateViewport(data, error, {viewboxX, viewboxY, renderFrame});
}
renderFrame();


const zoomSlider = document.getElementById('zoom-slider');




function save() {
  const output = `${editor.getValue()}\n---\n${JSON.stringify(data)}`;

  const filename = prompt('Filename?');

  const blob = new Blob([output], {type: 'text/plain'});
  const url = URL.createObjectURL(blob);

  const tempLink = document.createElement('a');
  document.body.appendChild(tempLink);
  tempLink.setAttribute('href', url);
  tempLink.setAttribute('download', `${filename}.bit.json`);
  tempLink.click();
}
window.save = save;







let viewboxX = 0;
let viewboxY = 0;
let viewboxZoom = 1;


// Allow the user to change the zoom level of the viewbox by moving the slider.
function zoomViewbox() {
  viewboxZoom = zoomSlider.value / 100;
  resizePanes(resizePosition);
}

zoomSlider.addEventListener('change', zoomViewbox);
zoomSlider.addEventListener('input', zoomViewbox);


// Adjust the position of the viewbox when the user drags around the svg canvas.
let moveOnSvg = false;
viewport.addEventListener('mousedown', event => {
  moveOnSvg = event.target.getAttribute('id') === 'viewport' ||
    event.target.getAttribute('id') === 'block' ||
    event.target.getAttribute('id') === 'wire';

  if (moveOnSvg) { // Deselect all gates if clicking on the viewport background or a block.
    data.Gates.forEach(i => {
      i.active = false;
    });
    renderFrame([]);
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
      s.xPosition = (s.xPosition || 0) + viewboxZoom * event.movementX;
      s.yPosition = (s.yPosition || 0) + viewboxZoom * event.movementY;
    });
    renderFrame([]);
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
  viewport.setAttribute('viewBox', `${viewboxX} ${viewboxY} ${viewboxZoom * viewport.clientWidth} ${viewboxZoom * viewport.clientHeight}`);

  // Rerender the viewport
  renderFrame([]);
}

resizeBar.addEventListener('mousemove', function(event) {
  if (event.buttons > 0) {
    resizePosition = event.screenX;
    resizePanes(resizePosition);
  }
});

// Initial svg size
resizePanes(resizePosition);
