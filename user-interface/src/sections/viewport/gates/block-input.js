export function insert(group) {
  group.append('path')
    .attr('fill', 'transparent')
    .attr('stroke', 'black')
    .attr('stroke-width', 2)

  return group
}

export function merge(group) {
  group.select('path')
    .attr('fill', d => d.active ? 'green' : 'red')
    .attr('d', `M0,0 H20 V20 H0 V0`);

  return group
}
