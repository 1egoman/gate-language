import * as d3 from "d3";

const GATE_WIDTH = 30;
const GATE_HEIGHT = 50;

export default function renderViewport(viewport) {
  const svg = d3.select(viewport);

  const blocks = svg.append('g')
    .attr('class', 'layer layer-blocks');

  const wires = svg.append('g')
    .attr('class', 'layer layer-wires');

  const gates = svg.append('g')
    .attr('class', 'layer layer-gates');

  return data => {
    const allGates = data.Gates, allOutputs = data.Outputs;

    // Link block inputs and outputs with the block they belong to.
    let blockTree = {};
    allGates.forEach(gate => {
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
    const allBlocks = [];
    for (let blockName in blockTree) {
      for (let invocationNumber in blockTree[blockName]) {
        const inputs = Object.values(blockTree[blockName][invocationNumber].inputs);
        const outputs = Object.values(blockTree[blockName][invocationNumber].outputs);
        allBlocks.push({
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
    window.allBlocks = allBlocks


    const gatesSelection = gates.selectAll('.gate').data(allGates);

    // Add a new gates when new data elements show up
    const gatesSelectionEnter = gatesSelection.enter()
      .append('g')
      .attr('class', 'gate')
      .on('click', function(d) {
        d.active = true;
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
          } else {
            return 'transparent';
          }
        })
        .attr('d', d => {
          switch (d.Type) {
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
          default:
            return `M0,0 H30 V50 H0 V0`;
          }
        })

    gatesSelection.exit().remove()



    function getGateInputPosititon(gate, inputNumber) {
      if (gate.Type === 'BLOCK_OUTPUT' || gate.Type === 'BLOCK_INPUT') {
        return {x: gate.xPosition + 10, y: gate.yPosition + 10};
      }

      const spacingBetweenInputs = GATE_WIDTH / gate.Inputs.length;
      const startPadding = spacingBetweenInputs / 2;
      return {
        x: gate.xPosition + startPadding + (spacingBetweenInputs * inputNumber),
        y: gate.yPosition + GATE_HEIGHT - 6,
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

    const wirePaths = {};
    function appendWirePath(id, x, y) {
      if (wirePaths[id]) {
        wirePaths[id] += `L${x},${y}`;
      } else {
        wirePaths[id] = `M${x},${y}`;
      }
    }

    allGates.forEach(gate => {
      gate.Inputs.forEach((input, ct) => {
        const {x, y} = getGateInputPosititon(gate, ct);
        appendWirePath(input.Id, x, y);
      });

      gate.Outputs.forEach((output, ct) => {
        const {x, y} = getGateOutputPosititon(gate, ct);
        appendWirePath(output.Id, x, y);
      });
    });

    allOutputs.forEach(wire => {
      appendWirePath(wire.Id, 0, 0);
    });


    const wiresSelection = wires.selectAll('.wire').data(
      Object.keys(wirePaths).map(k => ({id: k, data: wirePaths[k]}))
    );

    // Add a new wires when new data elements show up
    const wireEnterSelection = wiresSelection.enter().append('g').attr('class', 'wire');

    wireEnterSelection.append('path')
      .attr('fill', 'transparent')
      .attr('stroke', 'black')
      .attr('stroke-width', 2)
      .attr('data-wire-id', d => d.id)

    const wireMergeSelection = wireEnterSelection.merge(wiresSelection);
    wireMergeSelection.select('path')
      .attr('d', d => d.data)

    wiresSelection.exit().remove()



    const blocksSelection = blocks.selectAll('.block').data(allBlocks);
    blocksSelection.enter()
      .append('rect')
        .attr('class', 'block')
        .attr('fill', 'silver')
        .attr('id', 'block')
    .merge(blocksSelection)
      .attr('class', d => `block block-${d.name}`)
      .attr('x', d => d.upperLeftBound[0])
      .attr('y', d => d.upperLeftBound[1])
      .attr('width', d => d.lowerRightBound[0] - d.upperLeftBound[0])
      .attr('height', d => d.lowerRightBound[1] - d.upperLeftBound[1])

    blocksSelection.exit().remove()
  }
}
