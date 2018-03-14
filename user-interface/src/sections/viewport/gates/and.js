export function insert(group) {
  group.append('path')
    .attr('fill', 'transparent')
    .attr('stroke', 'black')
    .attr('stroke-width', 2)

  return group
}

export function merge(group) {
  group.select('path')
    .attr('fill', d => d.active ? 'green' : 'transparent')
    .attr('d', `M0,15 C0,6.71572875 6.71572875,0 15,0 C23.2842712,0 30,6.71572875 30,15 L30,45 L0,45 L0,15 Z`)

  return group
}
