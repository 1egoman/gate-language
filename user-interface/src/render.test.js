import sinon from 'sinon';
import assert from 'assert';
import renderViewport from './render';

const Gates = [
  {
    "Id": 1,
    "Type": "BUILTIN_FUNCTION",
    "Label": "toggle",
    "Inputs": [
      
    ],
    "Outputs": [
      {
        "Id": 1,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 0
  },
  {
    "Id": 2,
    "Type": "BLOCK_INPUT",
    "Label": "Input 0 into block invert invocation 1",
    "Inputs": [
      {
        "Id": 1,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "Outputs": [
      {
        "Id": 2,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 1
  },
  {
    "Id": 3,
    "Type": "BLOCK_INPUT",
    "Label": "Input 0 into block foo invocation 1",
    "Inputs": [
      {
        "Id": 2,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "Outputs": [
      {
        "Id": 3,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 2
  },
  {
    "Id": 4,
    "Type": "SOURCE",
    "Label": "",
    "Inputs": [
      
    ],
    "Outputs": [
      {
        "Id": 4,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 2
  },
  {
    "Id": 5,
    "Type": "AND",
    "Label": "",
    "Inputs": [
      {
        "Id": 3,
        "Desc": "",
        "Start": null,
        "End": null
      },
      {
        "Id": 4,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "Outputs": [
      {
        "Id": 5,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 2
  },
  {
    "Id": 6,
    "Type": "GROUND",
    "Label": "",
    "Inputs": [
      
    ],
    "Outputs": [
      {
        "Id": 6,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 2
  },
  {
    "Id": 7,
    "Type": "OR",
    "Label": "",
    "Inputs": [
      {
        "Id": 5,
        "Desc": "",
        "Start": null,
        "End": null
      },
      {
        "Id": 6,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "Outputs": [
      {
        "Id": 7,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 2
  },
  {
    "Id": 8,
    "Type": "BLOCK_OUTPUT",
    "Label": "Output 0 from block foo invocation 1",
    "Inputs": [
      {
        "Id": 7,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "Outputs": [
      {
        "Id": 8,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 2
  },
  {
    "Id": 9,
    "Type": "GROUND",
    "Label": "",
    "Inputs": [
      
    ],
    "Outputs": [
      {
        "Id": 9,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 1
  },
  {
    "Id": 10,
    "Type": "OR",
    "Label": "",
    "Inputs": [
      {
        "Id": 8,
        "Desc": "",
        "Start": null,
        "End": null
      },
      {
        "Id": 9,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "Outputs": [
      {
        "Id": 10,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 1
  },
  {
    "Id": 11,
    "Type": "BLOCK_OUTPUT",
    "Label": "Output 0 from block invert invocation 1",
    "Inputs": [
      {
        "Id": 10,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "Outputs": [
      {
        "Id": 11,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "CallingContext": 1
  },
  {
    "Id": 12,
    "Type": "BUILTIN_FUNCTION",
    "Label": "led",
    "Inputs": [
      {
        "Id": 11,
        "Desc": "",
        "Start": null,
        "End": null
      }
    ],
    "Outputs": [
      
    ],
    "CallingContext": 0
  },
];
const Wires = [
  { "Id": 1, "Desc": "", "Start": null, "End": null },
  { "Id": 2, "Desc": "", "Start": null, "End": null },
  { "Id": 2, "Desc": "", "Start": null, "End": null },
  { "Id": 3, "Desc": "", "Start": null, "End": null },
  { "Id": 3, "Desc": "", "Start": null, "End": null },
  { "Id": 4, "Desc": "", "Start": null, "End": null },
  { "Id": 5, "Desc": "", "Start": null, "End": null },
  { "Id": 6, "Desc": "", "Start": null, "End": null },
  { "Id": 7, "Desc": "", "Start": null, "End": null },
  { "Id": 7, "Desc": "", "Start": null, "End": null },
  { "Id": 7, "Desc": "", "Start": null, "End": null },
  { "Id": 8, "Desc": "", "Start": null, "End": null },
  { "Id": 8, "Desc": "", "Start": null, "End": null },
  { "Id": 8, "Desc": "", "Start": null, "End": null },
  { "Id": 9, "Desc": "", "Start": null, "End": null },
  { "Id": 10, "Desc": "", "Start": null, "End": null },
  { "Id": 10, "Desc": "", "Start": null, "End": null },
  { "Id": 10, "Desc": "", "Start": null, "End": null },
  { "Id": 11, "Desc": "", "Start": null, "End": null },
];
const Contexts = [
  {
    "Id": 1,
    "Name": "invert",
    "Depth": 1,
    "Parent": 0,
    "Children": [2]
  },
  {
    "Id": 2,
    "Name": "foo",
    "Depth": 2,
    "Parent": 1,
    "Children": [],
  },
];

describe('render', () => {
  let viewport;
  beforeEach(() => {
    viewport && viewport.remove();
    viewport = document.createElement('svg');
  });

  it('should render a sample gate network to the viewport', () => {
    const renderFrame = sinon.spy();

    const updateViewport = renderViewport(viewport);
    updateViewport(
      {Gates, Wires, Contexts, Outputs: []},
      null,
      {viewboxX: 0, viewboxY: 0, renderFrame},
    );

    assert.equal(viewport.children.length, 3);

    // Verify that two blocks were drawn
    assert.equal(viewport.children[0].className, 'layer layer-blocks');
    assert.equal(viewport.children[0].children.length, 2);

    // Verify that 11 wires were drawn
    assert.equal(viewport.children[1].className, 'layer layer-wires');
    assert.equal(viewport.children[1].children.length, 11);

    // Verify that 12 gates were drawn
    assert.equal(viewport.children[2].className, 'layer layer-gates');
    assert.equal(viewport.children[2].children.length, 12);
  });
});
