import renderViewport from './render';

export default function initializeViewport(element, server) {
  // Create a new viewport
  const updateViewport = renderViewport(element);

  let viewboxX = 0, viewboxY = 0;

  let gateState = null;

  function renderFrame(data, error, updatedGateIds) {
    data.Gates = data.Gates || []
    data.Wires = data.Wires || []
    data.Outputs = data.Outputs || []

    // Update error bar state
    if (error) {
      document.getElementById('error-bar').style.display = 'flex';
      document.getElementById('error-bar').innerText = error;
    } else {
      document.getElementById('error-bar').style.display = 'none';
    }

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

        if (!updates.Gates || !updates.Wires) { return; }
        if (updates.Gates.length === 0) { return; }

        // Update the state of each gate, and the powered state of each wire
        data.Gates.forEach((gate, index) => {
          data.Gates[index].State = updates.Gates[index].State;
        });
        data.Wires.forEach((wire, index) => {
          data.Wires[index].Powered = updates.Wires[index].Powered;
        });

        // Rerender the viewport.
        updateViewport(data, {viewboxX, viewboxY, renderFrame});
      }).catch(err => {
        renderFrame({}, err, []);
      });
    } else {
      // Even if there wasn't any change to what should be rendered on screen, Rerender the viewport
      // anyway.
      updateViewport(data, {viewboxX, viewboxY, renderFrame});
    }
  }

  return renderFrame;
}
