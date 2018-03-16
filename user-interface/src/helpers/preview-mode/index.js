import positionGates from '../../position-gates'; 

// When in preview mode, don't render an editor. Instead, connect over websockets to a server
// running on the local system and whenever a new ast update is pushed, update what is shown in the
// visualization.
export default function connectToPreviewWebsocket(renderFrame, websocketsServer) {
  const ws = new WebSocket(`${websocketsServer}/v1/websocket`);
  ws.onmessage = event => {
    const payload = JSON.parse(event.data);
    let data = payload,
        error = null;

    if (payload.Gates) {
      // Position gates on the screen
      data = positionGates(data);
    } else {
      error = payload.Error;
    }

    // Rerender using the data received.
    renderFrame(data, error, data.Gates ? data.Gates.map(i => i.Id) : []);
  }

  // On close, wait three seconds and try to connect again.
  ws.onclose = event => {
    setTimeout(connectToPreviewWebsocket, 3000);
  }
}
