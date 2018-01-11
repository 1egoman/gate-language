import CodeMirror from 'codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/monokai.css';
import 'codemirror/mode/jsx/jsx';
import 'codemirror/addon/mode/simple';

import './index.css';

import renderViewport from './render';

// import deepDiff from 'deep-diff';
import debounce from 'lodash.debounce';

import { generateBlocksFromGates } from './block-helpers';

import registerServiceWorker from './registerServiceWorker';
registerServiceWorker();



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
  const editor = CodeMirror(element, {
    lineNumbers: true,
    // value: `let a = toggle()\nled(a)`,
    // value: `let a = toggle()\nlet b = toggle()\nled(a and b)`,
//     value: `
// block foo(a) {
//   let result = ((a and 1) or 0)
//   return result
// }
//
// block bla(c) {
//   return (c and foo(c))
// }
//
// block invert(b) {
//   let v = foo(b)
//   let result = (v or 0)
//   return result
// }
// led(invert(toggle()))
//     
//     
//     
//     `,
    value: `
// A jk-flip-flop is a latch that takes two inputs - a and b.
// When a signal is provided to either input, the latch persists
// that state and outputs it on its only output, q.
block sr_latch(a b) {
  let nq = (not (a or q))
  let q = (not (b or nq))
  return q
}

// A jk-flip-flop that returns references to both states
// of the flip flop (ie, both q and not q)
block sr_latch_2(a b) {
  let nq = (not (a or q))
  let q = (not (b or nq))
  return q nq
}

block d_flipflop(d clk) {
  let a = ((d and clk))
  let b = (((not d) and clk))

  // Latch part of flip flop - this is a jk flip flop!
  let result = sr_latch(a b)
  return result
}

block d_flipflop_2(d clk) {
  let a = ((d and clk))
  let b = (((not d) and clk))

  // Latch part of flip flop - this is a jk flip flop!
  let q nq = sr_latch_2(a b)
  return q nq
}

block t_flipflop(t clock) {
  let switch = (t and clock)
  let q nq = sr_latch_2((nq and switch) (q and switch))
  return q
}

block main() {
  led(t_flipflop(1 toggle()))
  led(t_flipflop(1 toggle()))
  led(t_flipflop(1 toggle()))
}
main()
`,
    mode: 'bitlang',
    theme: 'monokai',
  });
  editor.setSize('100%', '100%');

  return editor;
}
const editor = createEditor(document.getElementById('editor-parent'));


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
    error = null;

    // Store data in global
    data = newData;
    data.Gates = data.Gates || []
    data.Wires = data.Wires || []
    data.Contexts = data.Contexts || []

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

      console.log('X', context.x, 'Y', context.y);
      console.log('Width', context.width, 'Height', context.height);
    });

    let globalGateCount = 0;
    data.Gates.forEach(gate => {
      const context = getContext(gate.CallingContext);
      if (context) {
        if (context.gateCount) {
          if (gate.Type === 'BLOCK_INPUT' || gate.Type === 'BLOCK_OUTPUT') {
            // All inputs and outputs are on the top border.
            gate.xPosition = context.x + (context.gateCount * 40)
            gate.yPosition = context.y
            context.gateCount += 1
          } else {
            // All the rest of the gates in a line below the inputs and outputs
            gate.xPosition = context.x + (context.gateCount * 40)
            gate.yPosition = context.y + 100
            context.gateCount += 1
          }
        } else {
          // Position the first gate in the lower right corner
          gate.xPosition = context.x
          gate.yPosition = context.y
          context.gateCount = 1
        }
      } else {
        // Position gates in the global scop in the upper left corner
        gate.xPosition = globalGateCount * 40
        gate.yPosition = 0
        globalGateCount += 1
      }
    });

    // Position each gate within its block
    /*
    blocks.forEach(block => {
      let gatePositionX = block.upperLeftBound[0],
          gatePositionY = block.upperLeftBound[1];

      // block.inputs.forEach(inp => {
      //   inp.xPosition = 0//gatePositionX;
      //   inp.yPosition = 0//gatePositionY;
      //   gatePositionX += 30;
      //   if (gatePositionX > block.lowerRightBound[0]) {
      //     gatePositionX = 0;
      //     gatePositionY += 100;
      //   }
      // });
      //
      // gatePositionX = block.upperLeftBound[0];
      // gatePositionY = block.lowerRightBound[1];
      // block.outputs.forEach(out => {
      //   out.xPosition = 0 //gatePositionX;
      //   out.yPosition = 0 //gatePositionY;
      //   gatePositionX += 30;
      //   if (gatePositionX > block.lowerRightBound[0]) {
      //     gatePositionX = 0;
      //     gatePositionY += 100;
      //   }
      // });

      gatePositionX = block.upperLeftBound[0];
      gatePositionY = block.upperLeftBound[1];
      [...block.contents, ...block.inputs, ...block.outputs].forEach(gate => {
        console.log('SETTING GATE', gate);
        gate.xPosition = gatePositionX;
        gate.yPosition = gatePositionY;
        gatePositionX += 40;

        // console.log(block.label, gate, gatePositionX, block.lowerRightBound[0])
        if (gatePositionX > block.lowerRightBound[0]) {
          gatePositionX = block.upperLeftBound[0];
          gatePositionY += 100;
        }
      });
      if (block.contents.length > 0) {
        block.contents[0].yPosition = block.lowerRightBound[1]
        block.outputs[block.outputs.length-1].xPosition = block.lowerRightBound[0]
      }
    });

    // Loop through each block, and if it doesn't already have an x or y position then it must be in
    // the global space.
    let gatePositionX = 0, gatePositionY = 500;
    data.Gates.forEach(gate => {
      if (gate.xPosition === undefined || gate.yPosition === undefined) {
        gate.xPosition = 0//gatePositionX
        gate.yPosition = 0//gatePositionY
        gate.xPosition = gatePositionX
        gate.yPosition = gatePositionY
        gatePositionX += 50
      }
    })
    */

    renderFrame(data.Gates);
  }).catch(err => {
    console.error(err.stack);
    // Set a global error variable
    error = err.message;
  });
}, 1000);

Object.defineProperty(window, 'gates', {
  get: function() {
    return data.Gates;
  },
  set: function(value) {
    data.Gates = value;
    renderFrame();
  },
})



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

// Update the powered state of any wires and redraw the viewport.
let gateState = null;
function renderFrame() {
  data.Gates = data.Gates || []
  data.Wires = data.Wires || []
  data.Outputs = data.Outputs || []

  function setWire(id, powered) {
    const wire = data.Wires.find(i => i.Id === id);
    if (wire) {
      wire.powered = powered;

      if (updatedWireIds.indexOf(id) === -1) {
        updatedWireIds.push(id);
      }
    }
  }

  function getWire(id) {
    const wire = data.Wires.find(i => i.Id === id);
    if (wire) {
      return wire.powered;
    }
  }

  // Update the powered state of the wires.
  const updatedWireIds = [];

  // Calculate a hash of the current gate's state
  const newGateState = JSON.stringify(
    data.Gates.filter(i => i.Type === 'BUILTIN_FUNCTION' && ['toggle', 'momentary'].indexOf(i.Label) !== -1)
      .map(i => [i.id, i.state])
      .sort((a, b) => b[0] - a[0])
      .map(i => i[1])
  ) + data.Gates.length.toString();

  // If the hash doesn't match the previous hash that was stored, recalculate the stat of all gates.
  if (gateState !== newGateState) {
    gateState = newGateState;

    // Update wire state.
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
  }

  // Rerender the viewport.
  updateViewport(data, error, {viewboxX, viewboxY, renderFrame});
}

// Initial frame render.
renderFrame();






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
let viewboxZoom = 1.5; // 1


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

// Initial svg size
resizePanes(resizePosition);
