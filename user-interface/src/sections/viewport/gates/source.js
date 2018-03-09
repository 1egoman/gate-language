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
    .attr('d', `M10,0 H20 V22 H30 V34 H20 V50 H10 V32 H0 V22 H10 V0`)

  return group
}
