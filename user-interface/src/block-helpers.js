const GATE_WIDTH = 30;
const GATE_HEIGHT = 50;
const BLOCK_PADDING = 10;

export function generateBlocksFromGates(gates) {
  // Link block inputs and outputs with the block they belong to.
  let blockTree = {};
  gates.forEach(gate => {
    // Don't care about gates that are at the lowest level - they aren't in a block, they are in the
    // root of the viewport. Don't need to surround them with a box.
    if (gate.CallingContext === 0) {
      return
    }

    if (gate.Type === 'BLOCK_INPUT' || gate.Type === 'BLOCK_OUTPUT') {
      const parts = gate.Label.match(/(?:Input|Output) ([0-9]+) (?:into|from) block (.+) invocation ([0-9]+)/);
      if (parts === null) {
        throw new Error(`Gate label for ${gate.Type} doesn't match the expected format.`);
      }

      const inputNumber = parseInt(parts[1], 10),
            blockName = parts[2],
            invocationNumber = parts[3];

      blockTree[gate.CallingContext] = blockTree[gate.CallingContext] || {
        label: `${blockName} id=${gate.CallingContext} invocation=${invocationNumber}`,
        inputs: {},
        outputs: {},
        contents: [],
      };

      if (gate.Type === 'BLOCK_INPUT') {
        blockTree[gate.CallingContext].inputs[inputNumber] = gate;
      } else {
        blockTree[gate.CallingContext].outputs[inputNumber] = gate;
      }
    } else {
      blockTree[gate.CallingContext] = blockTree[gate.CallingContext] || {
        label: null,
        inputs: {},
        outputs: {},
        contents: [],
      };

      blockTree[gate.CallingContext].contents.push(gate);
    }
  });

  // Convert block tree into an array of blocks
  let blocks = [];
  for (let blockCallingContext in blockTree) {
    const inputs = Object.values(blockTree[blockCallingContext].inputs);
    const outputs = Object.values(blockTree[blockCallingContext].outputs);
    const contents = blockTree[blockCallingContext].contents;
    blocks.push({
      id: blockCallingContext,
      label: blockTree[blockCallingContext].label,

      inputs,
      outputs,
      contents,

      // Calculate the upper left and lower right corners of the block rectangle.
      upperLeftBound: [
        Math.min.apply(Math, [
          ...inputs.map(i => i.xPosition + 10),
          ...outputs.map(i => i.xPosition + 10),
          ...contents.map(i => i.xPosition),
        ]) - BLOCK_PADDING,
        Math.min.apply(Math, [
          ...inputs.map(i => i.yPosition + 10),
          ...outputs.map(i => i.yPosition + 10),
          ...contents.map(i => i.yPosition),
        ]) - BLOCK_PADDING,
      ],
      lowerRightBound: [
        Math.max.apply(Math, [
          ...inputs.map(i => i.xPosition + 10),
          ...outputs.map(i => i.xPosition + 10),
          ...contents.map(i => i.xPosition + GATE_WIDTH),
        ]) + BLOCK_PADDING,
        Math.max.apply(Math, [
          ...inputs.map(i => i.yPosition + 10),
          ...outputs.map(i => i.yPosition + 10),
          ...contents.map(i => i.yPosition + GATE_HEIGHT),
        ]) + BLOCK_PADDING,
      ],
    });
  }

  // Add the parent of each block to each block.
  blocks = blocks.map(block => {
    return {...block, parent: getParentBlock(gates, blocks, block)};
  });

  // Add "stack depth" of block to the block too in the `depth` key.
  blocks = blocks.map(block => {
    let currentDepthBlock = block;
    let depth = 0;
    while (true) {
      currentDepthBlock = currentDepthBlock.parent
      if (currentDepthBlock) {
        depth += 1;
      } else {
        break;
      }
    }

    return {...block, depth};
  });

  const numberOfGatesInEachBlock = blocks.reduce((acc, block) => {
    // Add all gates in the current block to the correct entry.
    let numberOfGatesInBlock = block.contents.length + block.inputs.length + block.outputs.length;
    acc[block.id] = (acc[block.id] || 0) + numberOfGatesInBlock;

    // Add the count for all gates in the current block to all parent blocks too.
    let currentDepthBlock = block;
    while (true) {
      currentDepthBlock = currentDepthBlock.parent;
      if (currentDepthBlock) {
        acc[currentDepthBlock.id] = (acc[currentDepthBlock.id] || 0) + numberOfGatesInBlock;
      } else {
        break;
      }
    }
    return acc;
  }, {});

  return blocks.map(block => {
    return {...block, deepGateCount: numberOfGatesInEachBlock[block.id]};
  });
}

export function getParentBlock(gates, blocks, block) {
  const inputWire = block.inputs[0];
  // console.log('GET PARENT FOR', block)
  const gateThatConnectsExternally = gates.find(gate => {
    // Ensure that the gate we just found is not in `block`.
    if (block.id === gate.CallingContext.toString()) {
      return false;
    }

    // Look for a gate whose output connects to the inputWire or whose input connects to
    // the outputWire.
    // console.log(gate.Outputs.map(i => i.Id), inputWire.Id)
    // console.log(gate.Inputs.map(i => i.Id), inputWire.Id)
    return gate.Outputs.find(outp => outp && outp.Id === inputWire.Id) || gate.Inputs.find(inp => inp && inp.Id === inputWire.Id);
  });

  // console.log('gateThatConnectsExternally', gateThatConnectsExternally);

  if (gateThatConnectsExternally && gateThatConnectsExternally.CallingContext > 0) {
    // Figure out the position of the block that `gateThatConnectsExternally` is within.
    const parentBlock = blocks.find(b => b.id === gateThatConnectsExternally.CallingContext.toString());
    // console.log('PARENT BLOCK', parentBlock)
    return parentBlock;
  } else {
    // The gate's parent is calling context 0 - the global namespace.
    // console.log('PARENT BLOCK', null)
    return null;
  }
}
