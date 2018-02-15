import * as d3 from "d3";

import * as gatesSource from './gates/source';
import * as gatesGround from './gates/ground';
import * as gatesAnd from './gates/and';
import * as gatesOr from './gates/or';
import * as gatesNot from './gates/not';
import * as gatesBlockInput from './gates/block-input';
import * as gatesBlockOutput from './gates/block-output';
import * as gatesBuiltinMomentary from './gates/builtin-momentary';
import * as gatesBuiltinToggle from './gates/builtin-toggle';
import * as gatesBuiltinLed from './gates/builtin-led';
import * as gatesBuiltinTFlipFlop from './gates/builtin-tflipflop';

const GATE_RENDERERS = {
  'SOURCE': gatesSource,
  'GROUND': gatesGround,
  'AND': gatesAnd,
  'OR': gatesOr,
  'NOT': gatesNot,
  'BLOCK_INPUT': gatesBlockInput,
  'BLOCK_OUTPUT': gatesBlockOutput,
  'BUILTIN_FUNCTION': {
    'momentary': gatesBuiltinMomentary,
    'toggle': gatesBuiltinToggle,
    'led': gatesBuiltinLed,
    'tflipflop': gatesBuiltinTFlipFlop,
  },
};

const GATE_WIDTH = 30;
const GATE_HEIGHT = 50;
const BLOCK_PADDING = 10;

const BUILTIN_GATE_MOUSEDOWN_HANDLERS = {
  toggle(data) {
    data.State = data.State === 'on' ? 'off' : 'on';
  },
  momentary(data) {
    data.State = 'on';
  }
}

const BUILTIN_GATE_MOUSEUP_HANDLERS = {
  momentary(data) {
    data.State = 'off';
  },
}

function renderGates(gateGroup, {gates, wires, renderFrame}) {
  const gatesSelection = gateGroup.selectAll('.gate').data(gates);

  // Add a new gates when new data elements show up
  const gatesSelectionEnter = gatesSelection.enter()
    .append('g')
    .attr('class', 'gate')
    .on('click', function(d) {
      if (!d3.event.shiftKey) {
        // Clicking on a gate selects it.
        d.active = true;
        renderFrame([d.Id]);
      }
    })
    .on('mousedown', function(d) {
      if (d3.event.shiftKey && d.Type === 'BUILTIN_FUNCTION') {
        // If the builtin has a click handler, call it.
        const clickHandler = BUILTIN_GATE_MOUSEDOWN_HANDLERS[d.Label];
        if (clickHandler) {
          clickHandler(d);
          renderFrame([d.Id]);
        }
      }
    })
    .on('mouseup', function(d) {
      if (d3.event.shiftKey && d.Type === 'BUILTIN_FUNCTION') {
        // If the builtin has a mouseup handler, call it.
        const mouseupHandler = BUILTIN_GATE_MOUSEUP_HANDLERS[d.Label];
        if (mouseupHandler) {
          mouseupHandler(d);
          renderFrame([d.Id]);
        }
      }
    })
  gatesSelectionEnter.append('g')
    .attr('class', 'gate-contents')
  gatesSelectionEnter.append('text')
    .attr('fill', 'black')
    .attr('transform', 'translate(0,-5)')
    .attr('pointer-events', 'none')

  gatesSelectionEnter.select('.gate-contents')
    .attr('data-type', function(d) {
      return d.Type;
    })
    .call(function(selection) {
      selection.each(function(d, i) {
        let renderer = GATE_RENDERERS[d.Type];
        if (d.Type === 'BUILTIN_FUNCTION') {
          renderer = renderer[d.Label];
        }
        if (renderer) {
          renderer.insert(d3.select(this), d);
        } else {
          // Draw default gate shape
          d3.select(this).append('path');
        }
      });
    });

  const gatesMergeSelection = gatesSelectionEnter.merge(gatesSelection);
  gatesMergeSelection
    .attr('data-id', d => d.Id)
    .attr('transform', d => {
      return `translate(${d.xPosition || 0},${d.yPosition || 0})`;
    });

  // FIXME: help with the below. I'm hacking around d3 and I don't like it.
  gatesMergeSelection.select('.gate-contents')
    .attr('transform', d => {
      if (d.rotate === 180) {
        return `rotate(180, ${d.width/2} ${GATE_HEIGHT/2})`;
        } else {
          return '';
        }
    })
    .call(function(selection) {
      selection.each(function(d, i) {
        const elem = d3.select(this);
        const oldType = elem.attr('data-type') + ' ' + elem.attr('data-label');
        const newType = d.Type + ' ' + d.Label;

        // Find a renderer for the data item.
        let renderer = GATE_RENDERERS[d.Type];
        if (d.Type === 'BUILTIN_FUNCTION') {
          renderer = renderer[d.Label];
        }

        if (!renderer) {
          return
        }

        // If the item's type has changed, then delete everything inside of `.gate-contents` and
        // redraw it.
        if (oldType !== ' ' && oldType !== newType) {
          elem.selectAll('*').remove()
          renderer.insert(elem, d);
        }

        // Then, update the gate contents.
        renderer.merge(elem, d, {gates, wires});
      });
    })
    .attr('data-label', d => d.Label)
    .attr('data-type', d => d.Type)

  /* gatesMergeSelection.select('text') */
  /*   .text(d => `${d.Id} ${d.CallingContext}`); */

  gatesSelection.exit().remove()
}

function renderWires(wireGroup, {wires, gates, outputs, renderFrame}) {
  const wirePaths = {};
  function getGateInputPosititon(gate, inputNumber) {
    if (gate.Type === 'BLOCK_OUTPUT' || gate.Type === 'BLOCK_INPUT') {
      return {x: gate.xPosition + 10, y: gate.yPosition + 10};
    }
    if (gate.Type === 'BUILTIN_FUNCTION' && gate.Label === 'tflipflop') {
      if (inputNumber === 0) {
        // The first input, t, is right in the middle.
        return {
          x: gate.xPosition,
          y: gate.yPosition + (GATE_HEIGHT / 2),
        };
      } else if (inputNumber === 1) {
        // reset is at the bottom
        return {
          x: gate.xPosition,
          y: gate.yPosition + (GATE_HEIGHT * 0.75),
        };
      } else if (inputNumber === 2) {
        // set is at the top
        return {
          x: gate.xPosition,
          y: gate.yPosition + (GATE_HEIGHT * 0.25),
        };
      }
    }
    if (gate.Type === 'BUILTIN_FUNCTION' && gate.Label === 'led') {
      return {x: gate.xPosition + (GATE_WIDTH / 2), y: gate.yPosition + (GATE_WIDTH / 2)};
    }

    const spacingBetweenInputs = GATE_WIDTH / gate.Inputs.length;
    const startPadding = spacingBetweenInputs / 2;
    return {
      x: gate.xPosition + startPadding + (spacingBetweenInputs * inputNumber),
      y: gate.yPosition - (gate.rotate === 180 ? (gate.Type === 'NOT' ? 20 : 40) : 0) + (gate.Type === 'NOT' ? 40 : GATE_HEIGHT) - 6,
    }
  }

  function getGateOutputPosititon(gate, outputNumber) {
    if (gate.Type === 'BLOCK_OUTPUT' || gate.Type === 'BLOCK_INPUT') {
      return {x: gate.xPosition + 10, y: gate.yPosition + 10};
    }
    if (gate.Type === 'BUILTIN_FUNCTION' && gate.Label === 'tflipflop') {
      return {
        x: gate.xPosition + 80, /* 80 is the width of the t flip flop */
        y: gate.yPosition + ((GATE_HEIGHT / 4) * (outputNumber + 1)),
      };
    }

    const spacingBetweenOutputs = GATE_WIDTH / gate.Outputs.length;
    const startPadding = spacingBetweenOutputs / 2;
    return {
      x: gate.xPosition + startPadding + (spacingBetweenOutputs * outputNumber),
      y: gate.yPosition + (gate.rotate === 180 ? 50 : 0),
    }
  }

  function appendWirePath(id, x, y) {
    if (wirePaths[id]) {
      wirePaths[id] += `L${x},${y}`;
    } else {
      wirePaths[id] = `M${x},${y}`;
    }
  }

  gates.forEach(gate => {
    gate.Inputs.forEach((input, ct) => {
      const {x, y} = getGateInputPosititon(gate, ct);
      appendWirePath(input.Id, x, y);
    });

    gate.Outputs.forEach((output, ct) => {
      const {x, y} = getGateOutputPosititon(gate, ct);
      appendWirePath(output.Id, x, y);
    });
  });

  // All outputs conenct to to (0, 0)
  outputs.forEach(wire => {
    appendWirePath(wire.Id, 0, 0);
  });

  const wiresSelection = wireGroup.selectAll('.wire').data(
    Object.keys(wirePaths).map(k => ({
      id: k,
      data: wires.find(i => i.Id === parseInt(k, 10)),
      path: wirePaths[k],
    }))
  );

  // Add a new wires when new data elements show up
  const wireEnterSelection = wiresSelection.enter().append('g').attr('class', 'wire');

  wireEnterSelection.append('path')
    .attr('fill', 'transparent')
    .attr('id', 'wire')
    .attr('stroke', 'black')
    .attr('stroke-width', 2)
    .attr('data-wire-id', d => d.id)

  const wireMergeSelection = wireEnterSelection.merge(wiresSelection);
  wireMergeSelection
    .attr('data-id', d => d.Id)
  wireMergeSelection.select('path')
    .attr('d', d => d.path)
    .attr('stroke', d => {
      return d.data && d.data.Powered ? 'red' : 'black';
    })

  wiresSelection.exit().remove()
}

function renderContexts(contextGroup, {gates, contexts}) {
  // Resize contexts to match the size of the gates within
  (contexts || []).sort((a, b) => b.Depth - a.Depth).forEach(context => {
    const contextGates = gates.filter(i => i.CallingContext === context.Id);
    context.x = Math.min.apply(Math, [
      ...contextGates.map(i => {
        if (i.Type === 'BLOCK_INPUT' || i.Type === 'BLOCK_OUTPUT') {
          return i.xPosition + 10;
        } else {
          return i.xPosition;
        }
      }),
      ...context.Children.map(i => {
        const child = contexts.find(j => j.Id === i);
        return child.x;
      }),
    ]) - BLOCK_PADDING;
    context.width = Math.max.apply(Math, [
      ...contextGates.map(i => {
        if (i.Type === 'BLOCK_INPUT' || i.Type === 'BLOCK_OUTPUT') {
          return i.xPosition + 10;
        } else {
          return i.xPosition + GATE_WIDTH;
        }
      }),
      ...context.Children.map(i => {
        const child = contexts.find(j => j.Id === i);
        return child.x + child.width;
      }),
    ]) + BLOCK_PADDING - context.x;
    context.y = Math.min.apply(Math, [
      ...contextGates.map(i => {
        if (i.Type === 'BLOCK_INPUT' || i.Type === 'BLOCK_OUTPUT') {
          return i.yPosition + 10;
        } else {
          return i.yPosition;
        }
      }),
      ...context.Children.map(i => {
        const child = contexts.find(j => j.Id === i);
        return child.y;
      }),
    ]) - BLOCK_PADDING;
    context.height = Math.max.apply(Math, [
      ...contextGates.map(i => {
        if (i.Type === 'BLOCK_INPUT' || i.Type === 'BLOCK_OUTPUT') {
          return i.yPosition + 10;
        } else {
          return i.yPosition + GATE_HEIGHT;
        }
      }),
      ...context.Children.map(i => {
        const child = contexts.find(j => j.Id === i);
        return child.y + child.height;
      }),
    ]) + BLOCK_PADDING - context.y;
  });

  const blocksSelection = contextGroup.selectAll('.block').data(contexts || []);
  const blockEnterSelection = blocksSelection.enter();

  const blockEnterSelectionGroup = blockEnterSelection.append('g')
    .attr('class', 'block')
  blockEnterSelectionGroup.append('rect')
    .attr('id', 'block')
    .attr('fill', 'rgba(0, 0, 0, 0.1)')
  blockEnterSelectionGroup.append('text')
    .attr('id', 'block')
    .attr('fill', '#000')
    .attr('style', 'user-select: none;')

  const blockMergeSelection = blockEnterSelectionGroup.merge(blocksSelection);
  blockMergeSelection
    .attr('class', d => `block block-${(d.label || '').replace(/\s/g, '-').toLowerCase()}`)
    .attr('data-id', d => d.Id)
  blockMergeSelection.select('rect')
    .attr('x', d => d.x)
    .attr('y', d => d.y)
    .attr('width', d => d.width)
    .attr('height', d => d.height)
  blockMergeSelection.select('text')
    .attr('transform', d => `translate(${d.x},${d.y - 5})`)
    .text(d => `${d.Name} id=${d.Id} depth=${d.Depth}`)

  blocksSelection.exit().remove()
}

export default function renderViewport(viewport) {
  const svg = d3.select(viewport);

  const contexts = svg.append('g')
    .attr('class', 'layer layer-contexts');

  const wires = svg.append('g')
    .attr('class', 'layer layer-wires');

  const gates = svg.append('g')
    .attr('class', 'layer layer-gates');

  return (data, {viewboxX, viewboxY, renderFrame}) => {
    const allGates = data.Gates,
          allWires = data.Wires,
          allContexts = data.Contexts,
          allOutputs = data.Outputs;

    renderGates(gates, {gates: allGates, wires: allWires, renderFrame});
    renderWires(wires, {wires: allWires, gates: allGates, outputs: allOutputs, renderFrame});
    renderContexts(contexts, {wires: allWires, gates: allGates, contexts: allContexts});
  }
}
