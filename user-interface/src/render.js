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
    data.state = data.state === 'on' ? 'off' : 'on';
  },
  momentary(data) {
    data.state = 'on';
  }
}

const BUILTIN_GATE_MOUSEUP_HANDLERS = {
  momentary(data) {
    data.state = 'off';
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
      return d.Type;
    });

  const gatesMergeSelection = gatesSelectionEnter.merge(gatesSelection);
  gatesMergeSelection
    .attr('transform', d => {
      return `translate(${d.xPosition || 0},${d.yPosition || 0})`;
    });

  gatesMergeSelection.select('.gate-contents')
    .attr('data-type', function(d) {
      let renderer = GATE_RENDERERS[d.Type];
      if (d.Type === 'BUILTIN_FUNCTION') {
        renderer = renderer[d.Label];
      }
      if (renderer) {
        renderer.merge(d3.select(this), d, {gates, wires});
      } else {
        // Update default gate shape
        d3.select(this).select('path')
          .attr('fill', d => d.active ? 'green' : 'silver')
          .attr('d', `M0,0 H30 V50 H0 V0`);
      }
      return d.Type;
    });

  // gatesMergeSelection.select('text')
  //   .text(d => `${d.Id} ${d.CallingContext}`);

  gatesSelection.exit().remove()
}

function renderWires(wireGroup, {wires, gates, outputs, renderFrame}) {
  const wirePaths = {};
  function getGateInputPosititon(gate, inputNumber) {
    if (gate.Type === 'BLOCK_OUTPUT' || gate.Type === 'BLOCK_INPUT') {
      return {x: gate.xPosition + 10, y: gate.yPosition + 10};
    }
    if (gate.Type === 'BUILTIN_FUNCTION' || gate.Label === 'led') {
      return {x: gate.xPosition + (GATE_WIDTH / 2), y: gate.yPosition + (GATE_WIDTH / 2)};
    }

    const spacingBetweenInputs = GATE_WIDTH / gate.Inputs.length;
    const startPadding = spacingBetweenInputs / 2;
    return {
      x: gate.xPosition + startPadding + (spacingBetweenInputs * inputNumber),
      y: gate.yPosition + (gate.Type === 'NOT' ? 40 : GATE_HEIGHT) - 6,
    }
  }

  function getGateOutputPosititon(gate, outputNumber) {
    if (gate.Type === 'BLOCK_OUTPUT' || gate.Type === 'BLOCK_INPUT') {
      return {x: gate.xPosition + 10, y: gate.yPosition + 10};
    }

    const spacingBetweenOutputs = GATE_WIDTH / gate.Outputs.length;
    const startPadding = spacingBetweenOutputs / 2;
    return {
      x: gate.xPosition + startPadding + (spacingBetweenOutputs * outputNumber),
      y: gate.yPosition,
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
  wireMergeSelection.select('path')
    .attr('d', d => d.path)
    .attr('stroke', d => {
      return d.data && d.data.powered ? 'red' : 'black';
    })

  wiresSelection.exit().remove()
}

function renderBlocks(blockGroup, {gates, contexts}) {
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

  const blocksSelection = blockGroup.selectAll('.block').data(contexts || []);
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

  const blockMergeSelection = blocksSelection.merge(blocksSelection);
  blockMergeSelection
    .attr('class', d => `block block-${(d.label || '').replace(/\s/g, '-').toLowerCase()}`)
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

  const blocks = svg.append('g')
    .attr('class', 'layer layer-blocks');

  const wires = svg.append('g')
    .attr('class', 'layer layer-wires');

  const gates = svg.append('g')
    .attr('class', 'layer layer-gates');

  return (data, error, {viewboxX, viewboxY, renderFrame}) => {
    const allGates = data.Gates,
          allWires = data.Wires,
          allContexts = data.Contexts,
          allOutputs = data.Outputs;

    // If there's an error, render an error overlay.
    const errorOverlay = svg.selectAll('g#error-overlay').data(error ? [{error, viewboxX, viewboxY}] : []);
    const errorOverlayEnter = errorOverlay.enter()
      .append('g')
        .attr('id', 'error-overlay')

    errorOverlayEnter.append('rect')
      .attr('x', 0)
      .attr('y', 0)
      .attr('width', '100%')
      .attr('height', 100)
      .attr('fill', 'red')
    errorOverlayEnter.append('text')
      .attr('transform', `translate(20,50)`)
      .attr('fill', '#fff')

    const errorOverlayMerge = errorOverlayEnter.merge(errorOverlay);
    errorOverlayMerge
      .attr('transform', d => `translate(${d.viewboxX},${ d.viewboxY})`)
    errorOverlayMerge.select('text')
      .text(d => d.error);

    errorOverlay.exit().remove()


    renderGates(gates, {gates: allGates, wires: allWires, renderFrame});
    renderWires(wires, {wires: allWires, gates: allGates, outputs: allOutputs, renderFrame});
    renderBlocks(blocks, {wires: allWires, gates: allGates, contexts: allContexts});
  }
}
