import * as d3 from "d3";

const GATE_WIDTH = 30;
const GATE_HEIGHT = 50;

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

function renderGates(gateGroup, {gates, renderFrame}) {
  const gatesSelection = gateGroup.selectAll('.gate').data(gates);

  // Add a new gates when new data elements show up
  const gatesSelectionEnter = gatesSelection.enter()
    .append('g')
    .attr('class', 'gate')
    .on('click', function(d) {
      if (!d3.event.shiftKey) {
        // Clicking on a gate selects it.
        d.active = true;
        renderFrame([d]);
      }
    })
    .on('mousedown', function(d) {
      if (d3.event.shiftKey && d.Type === 'BUILTIN_FUNCTION') {
        // If the builtin has a click handler, call it.
        const clickHandler = BUILTIN_GATE_MOUSEDOWN_HANDLERS[d.Label];
        if (clickHandler) {
          clickHandler(d);
          renderFrame([d]);
        }
      }
    })
    .on('mouseup', function(d) {
      if (d3.event.shiftKey && d.Type === 'BUILTIN_FUNCTION') {
        // If the builtin has a mouseup handler, call it.
        const mouseupHandler = BUILTIN_GATE_MOUSEUP_HANDLERS[d.Label];
        if (mouseupHandler) {
          mouseupHandler(d);
          renderFrame([d]);
        }
      }
    })
  gatesSelectionEnter.append('path')
      .attr('fill', 'transparent')
      .attr('stroke', 'black')
      .attr('stroke-width', 2)

  gatesSelectionEnter.merge(gatesSelection)
    .attr('transform', d => {
      return `translate(${d.xPosition || 0},${d.yPosition || 0})`;
    })
    .select('path')
      .attr('stroke', d => d.active ? 'green' : 'black')
      .attr('fill', d => {
        if (d.active) {
          return 'green';
        } else if (d.Type === 'SOURCE') {
          return 'red';
        } else if (d.Type === 'GROUND') {
          return 'black';
        } else if (d.Type === 'BLOCK_INPUT') {
          return 'red';
        } else if (d.Type === 'BLOCK_OUTPUT') {
          return 'blue';
        } else if (d.state === 'on') {
          return 'magenta';
        } else {
          return 'transparent';
        }
      })
      .attr('d', d => {
        switch (d.Type) {
        case 'NOT':
          return `M17.2365571,9.47311425 L30,35 L0,35 L12.7634429,9.47311425
            C11.1248599,8.65222581 10,6.95747546 10,5 C10,2.23857625 12.2385763,0 15,0
            C17.7614237,0 20,2.23857625 20,5 C20,6.95747546 18.8751401,8.65222581
            17.2365571,9.47311425 Z`;
        case 'AND':
          return `M0,15 C0,6.71572875 6.71572875,0 15,0 C23.2842712,0
            30,6.71572875 30,15 L30,45 L0,45 L0,15 Z`;
        case 'OR':
          return `M29.9995379,44.8264114 C29.3110604,43.8025102 22.8584163,43 15,43
            C7.55891826,43 1.37826165,43.7195362 0.158268034,44.6652435 C0.128188016,44.6524168
            0.0986130759,44.6379035 0.0695437648,44.6217245 C-0.0916881233,44.7417332
            -0.134974039,44.8705056 -0.0603139832,45.0080419 C-0.0603139832,44.9468198
            -0.0397654827,44.8862377 0.000462099566,44.8264114 L0.0690459207,44.6214473
            C-2.07813531,43.4247556 -1.46553035,33.1366252 2.12962531,22.0718738
            C5.53049733,11.6050659 11.8324838,1.02434799 14.6220414,0.0576633394
            C14.7378215,-0.136928788 14.8847045,-0.266583994 15.0650959,-0.325196716
            C17.4041668,-1.08520694 24.4913873,10.3872278 28.1754469,21.7255972
            C31.8595064,33.0639667 32.4127184,43.5864935 30.0736474,44.3465038 C30.001938,44.3698036
            29.9283427,44.3836574 29.8529384,44.3882959 L29.9995379,44.8264114 Z`;
        case 'SOURCE':
          return `M10,0 H20 V22 H30 V34 H20 V50 H10 V32 H0 V22 H10 V0`;
        case 'BLOCK_INPUT':
        case 'BLOCK_OUTPUT':
          return `M0,0 H20 V20 H0 V0`;
        case 'BUILTIN_FUNCTION':
          switch (d.Label) {
          case 'toggle':
            if (d.state === 'on') {
              return `M0.5,0.5 L0.5,49.5 L29.5,49.5 L29.5,0.5 L0.5,0.5 Z M9,14.5 C5.96243388,14.5
              3.5,12.0375661 3.5,9 C3.5,5.96243388 5.96243388,3.5 9,3.5 C12.0375661,3.5
              14.5,5.96243388 14.5,9 C14.5,12.0375661 12.0375661,14.5 9,14.5 Z`;
            } else {
              // Circle on bottom
              return `M0.5,0.5 L0.5,49.5 L29.5,49.5 L29.5,0.5 L0.5,0.5 Z M21,14.5
              C17.9624339,14.5 15.5,12.0375661 15.5,9 C15.5,5.96243388 17.9624339,3.5 21,3.5
              C24.0375661,3.5 26.5,5.96243388 26.5,9 C26.5,12.0375661 24.0375661,14.5 21,14.5 Z`
            }
          case 'momentary':
            // 'M' with no circle
            return `M0.5,0.5 L0.5,49.5 L29.5,49.5 L29.5,0.5 L0.5,0.5 Z M14.6385883,7.38237665
            L18.3068605,3.71410446 L23.4263057,3.71410446 L23.4263057,18.4556447
            L19.0622929,18.4556447 L19.0622929,9.2995031 L14.5542,13.807596
            L10.0133185,9.26671443 L10.0133185,18.8044706 L6.5,18.8044706 L6.5,3.5
            L10.7562116,3.5 L14.6385883,7.38237665 Z`
          case 'led':
            return `M15,29.5 C23.0081289,29.5 29.5,23.0081289 29.5,15 C29.5,6.99187113
            23.0081289,0.5 15,0.5 C6.99187113,0.5 0.5,6.99187113 0.5,15 C0.5,23.0081289
            6.99187113,29.5 15,29.5 Z`;
          default:
            return `M0,0 H30 V50 H0 V0`;
          }
        default:
          return `M0,0 H30 V50 H0 V0`;
        }
      })

  gatesSelection.exit().remove()
}

function renderWires(wireGroup, {wires, gates, outputs, renderFrame}) {
  const wirePaths = {};
  function getGateInputPosititon(gate, inputNumber) {
    if (gate.Type === 'BLOCK_OUTPUT' || gate.Type === 'BLOCK_INPUT') {
      return {x: gate.xPosition + 10, y: gate.yPosition + 10};
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
  // console.log('allGates', allGates[2].Outputs[0])

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

function renderBlocks(blockGroup, {wires, gates}) {
  // Link block inputs and outputs with the block they belong to.
  let blockTree = {};
  gates.forEach(gate => {
    if (gate.Type === 'BLOCK_INPUT' || gate.Type === 'BLOCK_OUTPUT') {
      const parts = gate.Label.match(/(?:Input|Output) ([0-9]+) (?:into|from) block (.+) invocation ([0-9]+)/);
      if (parts === null) {
        throw new Error(`Gate label for ${gate.Type} doesn't match the default format.`);
      }

      const inputNumber = parseInt(parts[1], 10),
            blockName = parts[2],
            invocationNumber = parseInt(parts[3], 10);

      blockTree[blockName] = blockTree[blockName] || {};

      blockTree[blockName][invocationNumber] = blockTree[blockName][invocationNumber] || {
        inputs: {},
        outputs: {},
      };

      if (gate.Type === 'BLOCK_INPUT') {
        blockTree[blockName][invocationNumber].inputs[inputNumber] = gate;
      } else {
        blockTree[blockName][invocationNumber].outputs[inputNumber] = gate;
      }
    }
  });

  // Convert block tree into an array of blocks
  const blocks = [];
  for (let blockName in blockTree) {
    for (let invocationNumber in blockTree[blockName]) {
      const inputs = Object.values(blockTree[blockName][invocationNumber].inputs);
      const outputs = Object.values(blockTree[blockName][invocationNumber].outputs);
      blocks.push({
        name: blockName,
        invocationNumber,
        inputs,
        outputs,
        upperLeftBound: [
          Math.min.apply(Math, [...inputs.map(i => i.xPosition), ...outputs.map(i => i.xPosition)]),
          Math.min.apply(Math, [...inputs.map(i => i.yPosition), ...outputs.map(i => i.yPosition)]),
        ],
        lowerRightBound: [
          Math.max.apply(Math, [...inputs.map(i => i.xPosition), ...outputs.map(i => i.xPosition)]),
          Math.max.apply(Math, [...inputs.map(i => i.yPosition), ...outputs.map(i => i.yPosition)]),
        ],
      });
    }
  }

  const blocksSelection = blockGroup.selectAll('.block').data(blocks);
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
    .attr('class', d => `block block-${d.name}`)
  blockMergeSelection.select('rect')
    .attr('x', d => d.upperLeftBound[0])
    .attr('y', d => d.upperLeftBound[1])
    .attr('width', d => d.lowerRightBound[0] - d.upperLeftBound[0] + 20)
    .attr('height', d => d.lowerRightBound[1] - d.upperLeftBound[1] + 20)
  blockMergeSelection.select('text')
    .attr('transform', d => `translate(${d.upperLeftBound[0]},${d.upperLeftBound[1] - 5})`)
    .text(d => `${d.name} (#${d.invocationNumber})`)

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
    const allGates = data.Gates, allWires = data.Wires, allOutputs = data.Outputs;

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


    renderGates(gates, {gates: allGates, renderFrame});
    renderWires(wires, {wires: allWires, gates: allGates, outputs: allOutputs, renderFrame});
    renderBlocks(blocks, {wires: allWires, gates: allGates});

  }
}
