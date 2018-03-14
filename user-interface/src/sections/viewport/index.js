import renderViewport from './render';

const zoomSlider = document.getElementById('zoom-slider');

export default function initializeViewport(element, server) {
  // Create a new viewport
  const updateViewport = renderViewport(element);

  let viewboxX = 0, viewboxY = 0, viewboxZoom = 1.0;

  let gateState = null;

  // Store if the system is currently in an error state.
  let currentError = null;

  // Deselect any selected items when the svg is clicked.
  let moveOnSvg = false;
  let hoistedData = null;
  element.addEventListener('mousedown', event => {
    moveOnSvg = event.target.getAttribute('id') === 'viewport' ||
      event.target.getAttribute('id') === 'block' ||
      event.target.getAttribute('id') === 'wire';

    // Deselect all gates if clicking on the viewport background or a block.
    if (hoistedData && moveOnSvg) {
      hoistedData.Gates.forEach(i => {
        i.active = false;
      });
      renderFrame(hoistedData, currentError, []);
    }
  });
  element.addEventListener('mousemove', event => {
    if (!hoistedData) {
      return
    }

    let selected = hoistedData.Gates.filter(i => i.active === true);
    if (event.buttons > 0 && moveOnSvg) {
      viewboxX -= viewboxZoom * event.movementX;
      viewboxY -= viewboxZoom * event.movementY;
      updateViewport(hoistedData, {viewboxX, viewboxY, viewboxZoom, renderFrame});
    } else if (event.buttons > 0 && selected.length > 0) {
      selected.forEach(s => {
        s.xPosition = (s.xPosition || 0) + viewboxZoom * event.movementX;
        s.yPosition = (s.yPosition || 0) + viewboxZoom * event.movementY;
      });
      renderFrame(hoistedData, currentError, []);
    }
  });
  element.addEventListener('mouseup', event => {
    moveOnSvg = false;
  });

  // Control zooming of the viewport.

  const ZOOM_MINIMUM_LIMIT = 1; // Smallest zoom level possible.
  const ZOOM_MAXIMUM_LIMIT = 10; // Largest zoom level possible.

  // Used to scale the raw mouse event deltas so that one scroll wheel "click" zooms the correct
  // amount.
  const ZOOM_RATIO = 300;

  function updateZoom(zoomDelta) {
    viewboxZoom += zoomDelta;

    zoomSlider.value = viewboxZoom;

    // Stay within zoom limits
    if (viewboxZoom < ZOOM_MINIMUM_LIMIT) {
      viewboxZoom = ZOOM_MINIMUM_LIMIT;
    } else if (viewboxZoom > ZOOM_MAXIMUM_LIMIT) {
      viewboxZoom = ZOOM_MAXIMUM_LIMIT;
    } else {

      // If zooming within limits, then adjust the x and y posititons to ensure that the center of
      // the screen is maintained throught zooming.
      viewboxX -= zoomDelta * ZOOM_RATIO;
      viewboxY -= zoomDelta * ZOOM_RATIO;
    }

    // At this point, ZOOM_MINIMUM_LIMIT <= viewboxZoom <= ZOOM_MAXIMUM_LIMIT.
    updateViewport(hoistedData, {viewboxX, viewboxY, viewboxZoom, renderFrame});
  }

  // Option 1: zoom with the mouse wheel. This works well with a mouse wheel or a touchpad.
  element.addEventListener('wheel', event => {
    // Calculate zoom from scroll wheel posititon.
    const zoomDelta = -1 * event.wheelDelta / ZOOM_RATIO;
    updateZoom(zoomDelta);
  });

  // Option 2: zoom with the range slider. This is an option if the system you are on does not have
  // a mouse wheel or a touchpad.
  zoomSlider.addEventListener('input', event => {
    const delta = event.target.value - viewboxZoom;
    updateZoom(delta);
  });

  // (maybe) Future option 3: pinch to zoom?

  async function renderFrame(data, error, updatedGateIds) {
    data.Gates = data.Gates || []
    data.Wires = data.Wires || []
    data.Outputs = data.Outputs || []

    hoistedData = data;

    // Update error bar state
    currentError = error;
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

      try {
        const result = await window.fetch(`${server}/v1/run`, {
          method: 'POST',
          body: JSON.stringify(data),
          headers: {
            'Content-Type': 'text/plain',
            'Accept': 'application/json',
          },
        });

        if (!result.ok) {
          throw new Error(`Run failed: ${result.statusCode}`);
        }

        const updates = await result.json();

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
        updateViewport(data, {viewboxX, viewboxY, viewboxZoom, renderFrame});
      } catch (err) {
        renderFrame({}, err, []);
      }
    } else {
      // Even if there wasn't any change to what should be rendered on screen, Rerender the viewport
      // anyway.
      updateViewport(data, {viewboxX, viewboxY, viewboxZoom, renderFrame});
    }
  }

  return renderFrame;
}
