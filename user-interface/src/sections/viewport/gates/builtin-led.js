export function insert(group) {
  group.append('path')
    .attr('fill', 'transparent')
    .attr('stroke', 'black')
    .attr('stroke-width', 2)

  return group
}

export function merge(group) {
  group.select('path')
    .attr('fill', d => {
      if (d.active) {
        return 'green';
      } else if (d.State === 'on') {
        return 'magenta';
      } else {
        return 'silver';
      }
    })
    .attr('d', d => {
      return `M15,29.5 C23.0081289,29.5 29.5,23.0081289 29.5,15 C29.5,6.99187113
      23.0081289,0.5 15,0.5 C6.99187113,0.5 0.5,6.99187113 0.5,15 C0.5,23.0081289
      6.99187113,29.5 15,29.5 Z`;
    });

  return group
}
