import CodeMirror from 'codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/monokai.css';
import 'codemirror/mode/jsx/jsx';
import 'codemirror/addon/mode/simple';

import './index.css';

import renderViewport from './render';
import queryString from 'query-string';

// import deepDiff from 'deep-diff';
import debounce from 'lodash.debounce';

import registerServiceWorker from './registerServiceWorker';
registerServiceWorker();

const query = queryString.parse(window.location.search);

const server = query.server || 'http://localhost:8080';
const websocketsServer = (query.server || 'http://localhost:8080').replace('http', 'ws');

// When in preview mode, don't render an editor. Instead, connect over websockets to a server
// running on the local system and whenever a new ast update is pushed, update what is shown in the
// visualization.
const previewMode = Boolean(query.preview);
if (previewMode) {
  document.getElementById('resize-bar').style.display = 'none';
  document.getElementById('editor-parent').style.display = 'none';

  // Create websocket server connection
  function connect() {
    const ws = new WebSocket(`${websocketsServer}/v1/websocket`);
    ws.onmessage = function(event) {
      const payload = JSON.parse(event.data);
      if (payload.Gates) {
        data = payload;
        error = null;

        // Position gates on the screen
        data = positionGates(data);
      } else {
        error = payload.Error;
      }

      // Rerender using the data received.
      renderFrame(data.Gates.map(i => i.Id));
    }

    // On close, wait three seconds and try to connect again.
    ws.onclose = function(event) {
      setTimeout(connect, 3000);
    }
  }
  connect();
}


function createEditor(element) {
  // Define editor parameters
  CodeMirror.defineSimpleMode('bitlang', {
    // The start state contains the rules that are intially used
    start: [
      {regex: /(block)(\s+)([A-Za-z_][A-Za-z0-9_]*)/, token: ["keyword", null, "variable-2"]},
      {regex: /(let|return|block)\b/, token: "keyword"},
      {regex: /(?:1|0)/, token: "atom"},
      {regex: /\/\*/, token: "comment", next: "comment"},
      {regex: /\/\/[^\n]*/, token: "comment"},
      {regex: /(?:and|or|not|import)/, token: "property"},
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
  const editor = CodeMirror(element, {
    lineNumbers: true,
    value: ``,
    mode: 'bitlang',
    theme: 'monokai',
  });
  editor.setSize('100%', '100%');

  return editor;
}

// Given the response from a compilation, position all the gates on the screen.
function positionGates(data) {
  function getContext(id) {
    return data.Contexts.find(i => i.Id === id);
  }

  // Figure out all blocks that this gate network is made up of.
  const contextsSortedFromShallowestToDeepest = data.Contexts.sort((a, b) => a.Depth - b.Depth);

  // Position each block
  let rootContextX = 0, rootContextY = 0;
  contextsSortedFromShallowestToDeepest.forEach(context => {
    // Get parent and child contexts to the currently active context.
    const parent = getContext(context.Parent) || {};
    const childIndex = parent.Children ? parent.Children.findIndex(i => i === context.Id) : 0;
    const children = context.Children.map(getContext);

    context.x = (parent.x || rootContextX) + (context.Depth * (500 / 4)) + ((childIndex * 2) * 500)
    context.y = (parent.y || rootContextY) + (context.Depth * (500 / 4))
    context.width = 500 + (children.length * 500)
    context.height = 500 + (children.length * 500)

    // If in the root context, increment the position for hte next root block.
    if (!parent.x) {
      rootContextX += context.width;
      rootContextY += 0 //context.height;
    }
  });

  let spacingByContext = {};
  data.Gates.forEach(gate => {
    // Calculate the width of this gate.
    let gateWidth = 30;
    if (gate.Type === 'BLOCK_INPUT' || gate.Type === 'BLOCK_OUTPUT') {
      gateWidth = 20;
    }
    if (gate.Type === 'BUILTIN_FUNCTION' && gate.Label === 'tflipflop') {
      gateWidth = 80;
    }
    gate.width = gateWidth;

    const context = getContext(gate.CallingContext);
    if (context) {
      if (gate.Type === 'BLOCK_INPUT' || gate.Type === 'BLOCK_OUTPUT') {
        // All inputs are positioned on the left border, and all outputs on the right
        spacingByContext[context.Id] = spacingByContext[context.Id] || 0
        gate.xPosition = context.x + (gate.Type === 'BLOCK_OUTPUT' ? context.width : 0)
        gate.yPosition = context.y + spacingByContext[context.Id]
        context.gateCount += 1

        spacingByContext[context.Id] += gateWidth
      } else {
        // All the rest of the gates in a line below the inputs and outputs
        spacingByContext[context.Id] = spacingByContext[context.Id] || 0
        gate.xPosition = context.x + spacingByContext[context.Id]
        gate.yPosition = context.y + 100
        context.gateCount += 1
        spacingByContext[context.Id] += gateWidth
      }
    } else {
      // All the rest of the gates in a line below the inputs and outputs
      spacingByContext[null] = spacingByContext[null] || 0
      gate.xPosition = spacingByContext[null];
      gate.yPosition = 0
      spacingByContext[null] += gateWidth
    }
  });

  // Move inputs and outputs closer to the gates that they conenct to
  data.Gates.forEach(gate => {
    if (!(gate.Type === 'BLOCK_INPUT' || gate.Type === 'BLOCK_OUTPUT')) {
      return;
    }

    const context = getContext(gate.CallingContext);

    const gatesInContext = data.Gates
      .filter(i => i.Type === gate.Type && i.CallingContext.toString() === context.Id.toString());

    const positionOfGateInBlock = gatesInContext.findIndex(i => i.Id === gate.Id);

    if (gate.Type === 'BLOCK_OUTPUT') {
      gate.xPosition -= 40;
    }
    gate.yPosition = context.y + positionOfGateInBlock * 30;
  });

  // Move gates closer to their inputs and outputs
  data.Gates.forEach(gate => {
    if (gate.Inputs.length === 0 || gate.Type === 'BLOCK_INPUT' || gate.Type === 'BLOCK_OUTPUT') {
      return;
    }

    const gateConnectedToInput = data.Gates.find(g => {
      return g.Outputs.map(k => k.Id).indexOf(gate.Inputs[0].Id) !== -1;
    });

    if (gateConnectedToInput) {
      const wireLength = Math.sqrt(
        Math.pow(gateConnectedToInput.xPosition - gate.xPosition, 2),
        Math.pow(gateConnectedToInput.yPosition - gate.yPosition, 2),
      );

      if (wireLength > 100) {
        gate.xPosition = gateConnectedToInput.xPosition + 50;
        gate.yPosition = gateConnectedToInput.yPosition + 100;
      }
    }
  });

  // Final positioning step - make sure that gates don't intersect
  data.Gates.forEach((gate, index) => {
    if (gate.Type === 'BLOCK_INPUT' || gate.Type === 'BLOCK_OUTPUT') {
      return;
    }

    data.Gates.slice(index + 1)
      .filter(i => !(i.Type === 'BLOCK_INPUT' || i.Type === 'BLOCK_OUTPUT'))
      .filter(i =>  // Find other gates that intersect with this gate.
        i.xPosition >= gate.xPosition && i.xPosition <= gate.xPosition + 30 &&
        i.yPosition >= gate.yPosition && i.yPosition <= gate.yPosition + 30
      ).forEach((i, ct) => {
        i.xPosition += ((ct + 1) * 40) 
        // T flipflops are wider than normal gates, so add a bit of padding.
        if (gate.Type === 'BUILTIN_FUNCTION' && gate.Label === 'tflipflop') {
          i.xPosition += 60
        }
        i.yPosition += 10
      });
  });

  // Rotate gates to try to ensure that they are optimally placed
  data.Gates.forEach(gate => {
    if (gate.Outputs.length === 0 || gate.Type === 'BUILTIN_FUNCTION' || gate.Type === 'BLOCK_INPUT' || gate.Type === 'BLOCK_OUTPUT') {
      return;
    }

    const gateConnectedToInput = data.Gates.find(g => {
      return g.Inputs.map(k => k.Id).indexOf(gate.Outputs[0].Id) !== -1;
    });

    if (gateConnectedToInput && gateConnectedToInput.yPosition > gate.yPosition + 50) {
      gate.rotate = 180;
    } else {
      gate.rotate = 0;
    }
  });

  return data;
}

// If in normal editing mode, then render an editor
const editor = !previewMode ?
  createEditor(document.getElementById('editor-parent')) :
  null;


const compile = debounce(function compile(source) {
  return fetch(`${server}/v1/compile`, {
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
    error = null;

    // Store data in global
    data = newData;
    data.Gates = data.Gates || []
    data.Wires = data.Wires || []
    data.Contexts = data.Contexts || []

    data = positionGates(data)

    renderFrame(data.Gates.map(i => i.Id));
  }).catch(err => {
    console.error(err.stack);
    // Set a global error variable
    error = err.message;
    renderFrame([]);
  });
}, 1000);



let data = {Gates: [], Wires: [], Contexts: [], Outputs: []};
let error = null;

// When the user types in the editor, compile the source that they type and render it in the right
// pane.
if (editor) {
  editor.on('change', () => {
    const value = editor.getValue();
    localStorage.source = value;
    compile(value);
  });
  compile(editor.getValue());
}

// Get a reference to the svg viewport
const viewport = document.getElementById('viewport');
const updateViewport = renderViewport(viewport);

// Update the powered state of any wires and redraw the viewport.
let gateState = null;
function renderFrame(updatedGateIds) {
  data.Gates = data.Gates || []
  data.Wires = data.Wires || []
  data.Outputs = data.Outputs || []

  // Calculate a hash of the current gate's state
  const newGateState = JSON.stringify(
    data.Gates.filter(i => i.Type === 'BUILTIN_FUNCTION' && ['toggle', 'momentary'].indexOf(i.Label) !== -1)
      .map(i => [i.id, i.State])
      .sort((a, b) => b[0] - a[0])
      .map(i => i[1])
  ) + data.Gates.length.toString();

  // If the hash doesn't match the previous hash that was stored, recalculate the stat of all gates.
  if (gateState !== newGateState) {
    gateState = newGateState;

    return window.fetch(`${server}/v1/run`, {
      method: 'POST',
      body: JSON.stringify(data),
      headers: {
        'Content-Type': 'text/plain',
        'Accept': 'application/json',
      },
    }).then(result => {
      if (result.ok) {
        return result.json();
      } else {
        throw new Error(`Run failed: ${result.statusCode}`);
      }
    }).then(updates => {
      // Was an error received while compiling?
      if (updates.Error) {
        throw new Error(updates.Error);
      }

      if (updates.Gates.length === 0) {
        return;
      }

      // Update the state of each gate, and the powered state of each wire
      data.Gates.forEach((gate, index) => {
        data.Gates[index].State = updates.Gates[index].State;
      });
      data.Wires.forEach((wire, index) => {
        data.Wires[index].Powered = updates.Wires[index].Powered;
      });

      renderFrame(data.Gates.map(i => i.Id));
    }).catch(err => {
      console.error(err.stack);
      // Set a global error variable
      error = err.message;
      renderFrame([]);
    });
  }

  // Rerender the viewport.
  updateViewport(data, error, {viewboxX, viewboxY, renderFrame});
}

// Initial frame render.
renderFrame(data.Gates.map(i => i.Id));






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
const zoomSlider = document.getElementById('zoom-slider');
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

// When the browser window is resized, adjust the sizing of all the components on-screen.
function onViewportResize() {
  if (!previewMode) {
    // When in normal mode, start be resizing the panes.
    resizePanes(resizePosition);
  } else {
    // In preview mode, make the svg full-width.
    viewport.setAttribute('width', `${document.body.clientWidth}px`);
    viewport.setAttribute('height', `${document.body.clientHeight}px`);
  }
}
window.addEventListener('resize', onViewportResize);
onViewportResize();
