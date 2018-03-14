
// Given a context, return the x and y positions within the context for each gate.
function getGatePositionsWithinContext(context, gates) {
  const gatesInContext = gates.filter(i => i.CallingContext.toString() === context.Id.toString())

  return gatesInContext.map(gate => {
    return {
      x: gate.xPosition - context.x,
      y: gate.yPosition - context.y,
    };
  });
}

// Given the response from a compilation, position all the gates on the screen.
export default function positionGates(data) {
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
