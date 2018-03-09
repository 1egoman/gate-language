import positionGates from '../../position-gates';

export default function compile(server, source) {
  return fetch(`${server}/v1/compile`, {
    method: 'POST',
    body: source,
    headers: {
      'Content-Type': 'text/plain',
      'Accept': 'application/json',
    },
  }).then(result => {
    if (result.ok) {
      return result.json();
    } else {
      throw new Error(`Compilation failed: ${result.statusCode}`);
    }
  }).then(data => {

    // Was an error received while compiling?
    if (data.Error) {
      throw new Error(data.Error);
    }

    data.Gates = data.Gates || []
    data.Wires = data.Wires || []
    data.Contexts = data.Contexts || []

    data = positionGates(data)

    return data;
  });
};
