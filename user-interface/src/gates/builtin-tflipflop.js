function smallAnd(fg) {
  return fg.append('path')
    .attr('fill', 'black')
    .attr('d', `M3.83500004,12.835 L12.835041,12.835 L12.8356806,5.02607808 C12.709398,2.60813428 12.1606722,0.67925918 11.1993318,-0.766799173 C10.2676885,-2.16818666 9.31297295,-2.83500004 8.33500004,-2.83500004 C7.14878906,-2.83500004 6.1565574,-2.17358871 5.32654452,-0.786828217 C4.45851017,0.663457382 3.95786031,2.59990256 3.83500004,5 L3.83500004,12.835 Z`);
}
function smallOr(fg) {
  return fg.append('path')
    .attr('fill', 'black')
    .attr('d', `M8.42486328,-1.91109674 C6.55351974,-1.79933771 5.2204157,0.186900791 4.49960895,4.21455582 C3.82273172,7.99674561 3.59721021,11.075727 3.81859287,13.4445002 C5.35960006,12.6093844 6.89168771,12.1886399 8.41112764,12.1886399 C9.90653204,12.1886399 11.4141872,12.5961784 12.9305391,13.40518 C12.9645492,10.6038804 12.71139,7.37840972 12.1702316,3.73009811 C11.6034136,-0.0912016921 10.3387081,-1.9024798 8.42486328,-1.91109674 Z`);
}
function smallNot(fg) {
  return fg.append('path')
    .attr('fill', 'black')
    .attr('d', `M6.26321732,5.99244725 L1,9 L1,1 L6.26321732,4.00755275 C6.60793891,3.40558874 7.25660429,3 8,3 C9.1045695,3 10,3.8954305 10,5 C10,6.1045695 9.1045695,7 8,7 C7.25660429,7 6.60793891,6.59441126 6.26321732,5.99244725 Z`);
}

export function insert(group) {
  group.append('path')
    .attr('class', 'gate-tff-bg')
    .attr('fill', 'transparent')
    .attr('stroke', 'black')
    .attr('stroke-width', 2)

  const fg = group.append('g')
    .attr('class', 'gate-tff-fg');

  // All wires for the flip flop
  fg.append('path')
    .attr('class', 'w1')
    .attr('stroke', 'red')
    .attr('fill', 'transparent')
    .attr('d', 'M10,12 L5,12 L5,4 L70,4 L70,33 L65,33 L65,28 L35,20 L35,16 L37,16')

  fg.append('path')
    .attr('class', 'w2')
    .attr('stroke', 'red')
    .attr('fill', 'transparent')
    .attr('d', 'M10,35 L5,35 L5,43 L73,43 L73,13 L65,13 L65,18 L35,28 L35,32 L37,32')

  fg.append('path')
    .attr('class', 'w3')
    .attr('stroke', 'red')
    .attr('stroke-width', 2)
    .attr('fill', 'transparent')
    .attr('d', 'M10,32 L0,32')

  fg.append('path')
    .attr('class', 'w4')
    .attr('stroke', 'red')
    .attr('stroke-width', 2)
    .attr('fill', 'transparent')
    .attr('d', 'M10,15 L0,15')

  fg.append('path')
    .attr('class', 'w5')
    .attr('stroke', 'red')
    .attr('fill', 'transparent')
    .attr('d', 'M50,13 L56,13')

  fg.append('path')
    .attr('class', 'w6')
    .attr('stroke', 'red')
    .attr('fill', 'transparent')
    .attr('d', 'M50,34 L56,34')

  fg.append('path')
    .attr('class', 'w7')
    .attr('stroke', 'red')
    .attr('fill', 'transparent')
    .attr('d', 'M20,13 L38,11')

  fg.append('path')
    .attr('class', 'w8')
    .attr('stroke', 'red')
    .attr('fill', 'transparent')
    .attr('d', 'M20,34 L38,36')

  // All gates for the flip flop
  smallAnd(fg).attr('transform', 'translate(20, 5) rotate(90)');
  smallAnd(fg).attr('transform', 'translate(20, 25) rotate(90)');

  smallOr(fg).attr('transform', 'translate(50, 5) rotate(90)');
  smallOr(fg).attr('transform', 'translate(50, 25) rotate(90)');

  smallNot(fg).attr('transform', 'translate(55, 8)');
  smallNot(fg).attr('transform', 'translate(55, 28)');

  return group
}

export function merge(group, d, {gates, wires}) {
  group.select('.gate-tff-bg')
    .attr('fill', d => {
      if (d.active) {
        return 'green';
      } else {
        return 'silver';
      }
    })
    .attr('d', d => {
      return `M0,0 H80 V50 H0 V0`;
    });

  // The gates rendered on top
  const fg = group.select('.gate-tff-fg');

  const inputWires = d.Inputs.map(i => wires.find(j => i.Id === j.Id));
  const powered = inputWires[0].powered;

  fg.select('.w3').attr('stroke', powered ? 'red' : 'black'); // on if t is on
  fg.select('.w4').attr('stroke', powered ? 'red' : 'black'); // on if t is on

  if (d.State[1] === '1') {
    fg.select('.w1').attr('stroke', 'red');
    fg.select('.w2').attr('stroke', 'black');
    fg.select('.w5').attr('stroke', 'red');
    fg.select('.w6').attr('stroke', 'black');
    fg.select('.w7').attr('stroke', powered ? 'red' : 'black'); // on if t is on
    fg.select('.w8').attr('stroke', 'black'); // on if t is on
  } else {
    fg.select('.w1').attr('stroke', 'black');
    fg.select('.w2').attr('stroke', 'red');
    fg.select('.w5').attr('stroke', 'black');
    fg.select('.w6').attr('stroke', 'red');
    fg.select('.w7').attr('stroke', 'black'); // on if t is on
    fg.select('.w8').attr('stroke', powered ? 'red' : 'black'); // on if t is on
  }

  return group
}
