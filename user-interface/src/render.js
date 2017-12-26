import * as d3 from "d3";

const GATE_WIDTH = 30;
const GATE_HEIGHT = 50;

export default function renderViewport(viewport) {
  const svg = d3.select(viewport);

  const gates = svg.append('g')
    .attr('class', 'layer layer-gates');

  const wires = svg.append('g')
    .attr('class', 'layer layer-wires');

  return data => {
    const allGates = data.Gates, allWires = data.Wires;

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
          default:
            return `M0,0 H30 V50 H0 V0`;
          }
        })

    gatesSelection.exit().remove()




    const wireEnds = allGates.reduce((acc, gate) => {
      return [
        ...acc,
        ...gate.Inputs.map(inp => ({...inp, gate})),
        ...gate.Outputs.map(out => ({...out, gate})),
      ];
    }, []);

    const wirePaths = wireEnds.reduce((acc, i) => {
      let index = acc.findIndex(j => j.Id === i.Id);
      if (index !== -1) {
        acc[index].ends.push(i.gate);
        return acc;
      } else {
        return [...acc, {...i, gate: undefined, ends: [i.gate]}];
      }
    }, []);

    const wiresToDraw = wirePaths.reduce((acc, i) => {
      return [...acc, ...i.ends.reduce((acc, j, ct) => {
        if (ct < (i.ends.length - 1)) {
          return [...acc, {id: i.Id, start: i.ends[ct], end: i.ends[ct+1]}];
        } else {
          return acc;
        }
      }, [])];
    });

    if (!window.didPrint) {
      console.log('wireEnds', wireEnds);
      console.log('wirePaths', wirePaths);
      console.log('wiresToDraw', wiresToDraw);
      window.didPrint = true
    }

    const wiresSelection = wires.selectAll('.wire').data(wiresToDraw);

    // Add a new wires when new data elements show up
    const wireEnterSelection = wiresSelection.enter().append('g').attr('class', 'wire');

    wireEnterSelection.append('line')
      .attr('fill', 'transparent')
      .attr('stroke', 'black')
      .attr('stroke-width', 2)

    wireEnterSelection.append('text')

    const wireMergeSelection = wireEnterSelection.merge(wiresSelection);
    wireMergeSelection.select('line')
      .attr('x1', d => {
        // const spaceBetweenInputs = GATE_WIDTH / (d.start.Inputs.length + 2);
        // return d.start.xPosition + (spaceBetweenInputs * d.start.Inputs.findIndex(i => i.Id === d.id));
        return d.start.xPosition;
      })
      .attr('y1', d => {
        if (d.start.Inputs.find(i => i.Id === d.id)) {
          // If the wire is an input, render it at the top of the gate.
          return d.start.yPosition;
        } else {
          // If the wire is an output, render it at the bottom of the gate.
          return d.start.yPosition + GATE_HEIGHT;
        }
      })
      .attr('x2', d => d.end.xPosition)
      .attr('y2', d => {
        if (d.start.Outputs.find(i => i.Id === d.id)) {
          // If the wire is an input, render it at the top of the gate.
          return d.end.yPosition;
        } else {
          // If the wire is an output, render it at the bottom of the gate.
          return d.end.yPosition + GATE_HEIGHT;
        }
      })

    wireMergeSelection.select('text')
      .text(d => d.id)
      .attr('transform', d => `translate(${-1 * (d.end.xPosition - d.start.xPosition) / 2},${-1 * (d.end.yPosition - d.start.yPosition) / 2})`)

    wiresSelection.exit().remove()
  }
}
