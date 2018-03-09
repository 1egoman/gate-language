import './index.css';

import queryString from 'query-string';

import initializeEditor from './sections/editor/index';
import initializeViewport from './sections/viewport/index';

import './sections/pane-splits/index';

import registerServiceWorker from './registerServiceWorker';
registerServiceWorker();

const DEFAULT_SERVER = window.location.href.match(/^https?:\/\/lovelace-preview/) ?
  'https://lovelace-cloud.herokuapp.com' : 'http://localhost:8080';

const query = queryString.parse(window.location.search);

const server = query.server || DEFAULT_SERVER;
// const websocketsServer = (query.server || 'http://localhost:8080').replace('http', 'ws');

// When in preview mode, don't render an editor. Instead, connect over websockets to a server
// running on the local system and whenever a new ast update is pushed, update what is shown in the
// visualization.
// const previewMode = Boolean(query.preview);
// if (previewMode) {
//   document.getElementById('resize-bar').style.display = 'none';
//   document.getElementById('editor-parent').style.display = 'none';
//
//   // Create websocket server connection
//   function connect() {
//     const ws = new WebSocket(`${websocketsServer}/v1/websocket`);
//     ws.onmessage = function(event) {
//       const payload = JSON.parse(event.data);
//       if (payload.Gates) {
//         data = payload;
//         error = null;
//
//         // Position gates on the screen
//         data = positionGates(data);
//       } else {
//         error = payload.Error;
//       }
//
//       // Rerender using the data received.
//       renderFrame(data.Gates.map(i => i.Id));
//     }
//
//     // On close, wait three seconds and try to connect again.
//     ws.onclose = function(event) {
//       setTimeout(connect, 3000);
//     }
//   }
//   connect();
// }

// If in normal editing mode, then render an editor
// const editor = !previewMode ?
//   initializeEditor(document.getElementById('editor-parent'), server) :
//   null;

// Get a reference to the svg viewport
const renderFrame = initializeViewport(document.getElementById('viewport'), server);

const editor = initializeEditor(document.getElementById('editor-parent'), renderFrame, server);

console.log(editor)






// function save() {
//   const output = `${editor.getValue()}\n---\n${JSON.stringify(data)}`;
//
//   const filename = prompt('Filename?');
//
//   const blob = new Blob([output], {type: 'text/plain'});
//   const url = URL.createObjectURL(blob);
//
//   const tempLink = document.createElement('a');
//   document.body.appendChild(tempLink);
//   tempLink.setAttribute('href', url);
//   tempLink.setAttribute('download', `${filename}.bit.json`);
//   tempLink.click();
// }
// window.save = save;
//
//
//
//
//
//
//
// let viewboxX = 0;
// let viewboxY = 0;
// let viewboxZoom = 1;
//
//
// // Allow the user to change the zoom level of the viewbox by moving the slider.
// const zoomSlider = document.getElementById('zoom-slider');
// function zoomViewbox() {
//   viewboxZoom = zoomSlider.value / 100;
// }
//
// zoomSlider.addEventListener('change', zoomViewbox);
// zoomSlider.addEventListener('input', zoomViewbox);
//
//
// // Adjust the position of the viewbox when the user drags around the svg canvas.
// let moveOnSvg = false;
// viewport.addEventListener('mousedown', event => {
//   moveOnSvg = event.target.getAttribute('id') === 'viewport' ||
//     event.target.getAttribute('id') === 'block' ||
//     event.target.getAttribute('id') === 'wire';
//
//   if (moveOnSvg) { // Deselect all gates if clicking on the viewport background or a block.
//     data.Gates.forEach(i => {
//       i.active = false;
//     });
//     renderFrame([]);
//   }
// });
// viewport.addEventListener('mousemove', event => {
//   let selected = data.Gates.filter(i => i.active === true);
//   if (event.buttons > 0 && moveOnSvg) {
//     viewboxX -= viewboxZoom * event.movementX;
//     viewboxY -= viewboxZoom * event.movementY;
//   } else if (event.buttons > 0 && selected.length > 0) {
//     selected.forEach(s => {
//       s.xPosition = (s.xPosition || 0) + viewboxZoom * event.movementX;
//       s.yPosition = (s.yPosition || 0) + viewboxZoom * event.movementY;
//     });
//     renderFrame([]);
//   }
// });
